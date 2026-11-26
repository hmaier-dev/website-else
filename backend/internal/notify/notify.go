// Package notify verschickt optionale E-Mail-Benachrichtigungen
// (neue Kontaktanfrage, neue bezahlte Bestellung).
//
// Ohne SMTP_HOST ist alles aus – dann gibt es nur die Übersicht unter /manage.
//
// Umgebungsvariablen:
//
//	SMTP_HOST       z. B. smtp.example.org             (Pflicht, sonst aus)
//	SMTP_PORT       587 (STARTTLS, Standard) oder 465 (TLS)
//	SMTP_USERNAME   Login beim Mailanbieter
//	SMTP_PASSWORD   Passwort / App-Passwort
//	SMTP_FROM       Absenderadresse, z. B. website@elisa-adam.com
//	NOTIFY_EMAIL    Empfänger, mehrere mit Komma getrennt
//	SITE_URL        für Links in der Mail (Standard https://elisa-adam.com)
//	SMTP_INSECURE   "true" erlaubt Versand ohne Verschlüsselung (nur lokal zum Testen!)
package notify

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type Mailer struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
	SiteURL  string
	Insecure bool
	Timeout  time.Duration
}

// FromEnv liefert nil, wenn keine Benachrichtigung konfiguriert ist.
func FromEnv(getenv func(string) string) (*Mailer, error) {
	host := strings.TrimSpace(getenv("SMTP_HOST"))
	if host == "" {
		log.Println("notify: SMTP_HOST nicht gesetzt – keine E-Mail-Benachrichtigungen.")
		return nil, nil
	}
	m := &Mailer{
		Host:     host,
		Port:     587,
		Username: getenv("SMTP_USERNAME"),
		Password: getenv("SMTP_PASSWORD"),
		From:     strings.TrimSpace(getenv("SMTP_FROM")),
		SiteURL:  strings.TrimRight(getenv("SITE_URL"), "/"),
		Insecure: getenv("SMTP_INSECURE") == "true",
		Timeout:  30 * time.Second,
	}
	if p := getenv("SMTP_PORT"); p != "" {
		port, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("notify: SMTP_PORT %q ist keine Zahl", p)
		}
		m.Port = port
	}
	if m.SiteURL == "" {
		m.SiteURL = "https://elisa-adam.com"
	}
	for _, to := range strings.Split(getenv("NOTIFY_EMAIL"), ",") {
		if to = strings.TrimSpace(to); to != "" {
			if _, err := mail.ParseAddress(to); err != nil {
				return nil, fmt.Errorf("notify: NOTIFY_EMAIL enthält eine ungültige Adresse %q", to)
			}
			m.To = append(m.To, to)
		}
	}
	if len(m.To) == 0 {
		return nil, errors.New("notify: SMTP_HOST ist gesetzt, aber NOTIFY_EMAIL fehlt")
	}
	if m.From == "" {
		m.From = m.Username
	}
	if _, err := mail.ParseAddress(m.From); err != nil {
		return nil, errors.New("notify: SMTP_FROM fehlt oder ist ungültig")
	}
	log.Printf("notify: E-Mail-Benachrichtigungen an %d Empfänger über %s:%d", len(m.To), m.Host, m.Port)
	return m, nil
}

// oneLine entfernt Zeilenumbrüche, damit niemand über Formularfelder Mail-Header einschleusen kann.
func oneLine(s string) string {
	return strings.Join(strings.Fields(strings.NewReplacer("\r", " ", "\n", " ").Replace(s)), " ")
}

// build setzt die komplette E-Mail zusammen (Header + Text als quoted-printable).
func (m *Mailer) build(subject, body, replyTo, replyName string) ([]byte, error) {
	var buf bytes.Buffer
	from := mail.Address{Name: "Website Einfach Else", Address: m.From}
	h := func(k, v string) { fmt.Fprintf(&buf, "%s: %s\r\n", k, v) }
	h("From", from.String())
	h("To", strings.Join(m.To, ", "))
	if replyTo != "" {
		h("Reply-To", (&mail.Address{Name: oneLine(replyName), Address: oneLine(replyTo)}).String())
	}
	h("Subject", mime.QEncoding.Encode("utf-8", oneLine(subject)))
	h("Date", time.Now().Format(time.RFC1123Z))
	id := make([]byte, 12)
	rand.Read(id)
	domain := m.From[strings.LastIndex(m.From, "@")+1:]
	h("Message-ID", fmt.Sprintf("<%s@%s>", hex.EncodeToString(id), domain))
	h("MIME-Version", "1.0")
	h("Content-Type", "text/plain; charset=utf-8")
	h("Content-Transfer-Encoding", "quoted-printable")
	buf.WriteString("\r\n")
	qp := quotedprintable.NewWriter(&buf)
	if _, err := qp.Write([]byte(strings.ReplaceAll(body, "\n", "\r\n"))); err != nil {
		return nil, err
	}
	if err := qp.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Send verschickt eine Mail und wartet auf das Ergebnis.
func (m *Mailer) Send(ctx context.Context, subject, body, replyTo, replyName string) error {
	msg, err := m.build(subject, body, replyTo, replyName)
	if err != nil {
		return err
	}
	addr := net.JoinHostPort(m.Host, strconv.Itoa(m.Port))
	dialer := &net.Dialer{Timeout: m.Timeout}
	tlsConf := &tls.Config{ServerName: m.Host, MinVersion: tls.VersionTLS12}

	var conn net.Conn
	if m.Port == 465 {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConf)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("verbinden mit %s: %w", addr, err)
	}
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}
	c, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()

	if m.Port != 465 {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(tlsConf); err != nil {
				return fmt.Errorf("STARTTLS: %w", err)
			}
		} else if !m.Insecure {
			return errors.New("Mailserver bietet kein STARTTLS an – Versand abgebrochen (SMTP_INSECURE=true nur für lokale Tests)")
		}
	}
	if m.Username != "" {
		if err := c.Auth(smtp.PlainAuth("", m.Username, m.Password, m.Host)); err != nil {
			return fmt.Errorf("Anmeldung: %w", err)
		}
	}
	if err := c.Mail(m.From); err != nil {
		return err
	}
	for _, to := range m.To {
		if err := c.Rcpt(to); err != nil {
			return fmt.Errorf("Empfänger %s: %w", to, err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

// sendAsync verschickt im Hintergrund, damit das Formular nicht auf den Mailserver warten muss.
// Fehler werden nur geloggt – die Nachricht ist ja bereits gespeichert.
func (m *Mailer) sendAsync(what, subject, body, replyTo, replyName string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
		defer cancel()
		if err := m.Send(ctx, subject, body, replyTo, replyName); err != nil {
			log.Printf("notify: E-Mail (%s) konnte nicht verschickt werden: %v", what, err)
			return
		}
		log.Printf("notify: E-Mail (%s) verschickt", what)
	}()
}

// NewContactMessage: neue Anfrage über das Kontaktformular.
// Reply-To ist die Absenderin – „Antworten“ im Mailprogramm geht also direkt an sie.
func (m *Mailer) NewContactMessage(name, email, message string) {
	body := fmt.Sprintf("Neue Nachricht über das Kontaktformular\n\nVon:    %s\nE-Mail: %s\n\n%s\n\n—\nAlle Nachrichten: %s/manage/#nachrichten\n(Einfach auf diese Mail antworten, um %s zu schreiben.)\n",
		name, email, message, m.SiteURL, name)
	m.sendAsync("Kontaktanfrage", "Neue Anfrage von "+name, body, email, name)
}

// Order ist das, was in der Bestell-Mail steht.
type Order struct {
	Number  string
	Total   string
	Lines   []string
	Address string
	Email   string
	Name    string
}

// NewOrder: eine Bestellung wurde bezahlt.
func (m *Mailer) NewOrder(o Order) {
	body := fmt.Sprintf("Neue Bestellung %s ist bezahlt: %s\n\n%s\n\nLieferadresse:\n%s\n\nE-Mail: %s\n\n—\nAlle Bestellungen: %s/manage/#bestellungen\n",
		o.Number, o.Total, "- "+strings.Join(o.Lines, "\n- "), o.Address, o.Email, m.SiteURL)
	m.sendAsync("Bestellung "+o.Number, fmt.Sprintf("Neue Bestellung %s (%s)", o.Number, o.Total), body, o.Email, o.Name)
}
