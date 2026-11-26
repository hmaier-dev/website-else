package notify

import (
	"bufio"
	"context"
	"io"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeSMTP ist ein minimaler SMTP-Server ohne TLS, der die empfangene Mail speichert.
type fakeSMTP struct {
	ln       net.Listener
	mu       sync.Mutex
	data     string
	rcpt     []string
	authUsed bool
	done     chan struct{}
}

func startFake(t *testing.T) *fakeSMTP {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeSMTP{ln: ln, done: make(chan struct{}, 10)}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go f.handle(c)
		}
	}()
	return f
}

func (f *fakeSMTP) port() int { return f.ln.Addr().(*net.TCPAddr).Port }

func (f *fakeSMTP) handle(c net.Conn) {
	defer c.Close()
	r := bufio.NewReader(c)
	say := func(s string) { io.WriteString(c, s+"\r\n") }
	say("220 fake ESMTP")
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(cmd, "EHLO"):
			say("250-fake")
			say("250 AUTH PLAIN")
		case strings.HasPrefix(cmd, "AUTH"):
			f.mu.Lock()
			f.authUsed = true
			f.mu.Unlock()
			say("235 ok")
		case strings.HasPrefix(cmd, "MAIL"):
			say("250 ok")
		case strings.HasPrefix(cmd, "RCPT"):
			f.mu.Lock()
			f.rcpt = append(f.rcpt, strings.TrimSpace(line[8:]))
			f.mu.Unlock()
			say("250 ok")
		case cmd == "DATA":
			say("354 go")
			var b strings.Builder
			for {
				l, err := r.ReadString('\n')
				if err != nil {
					return
				}
				if l == ".\r\n" {
					break
				}
				b.WriteString(l)
			}
			f.mu.Lock()
			f.data = b.String()
			f.mu.Unlock()
			say("250 queued")
			f.done <- struct{}{}
		case cmd == "QUIT":
			say("221 bye")
			return
		default:
			say("250 ok")
		}
	}
}

func (f *fakeSMTP) wait(t *testing.T) *mail.Message {
	t.Helper()
	select {
	case <-f.done:
	case <-time.After(5 * time.Second):
		t.Fatal("keine Mail angekommen")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	msg, err := mail.ReadMessage(strings.NewReader(f.data))
	if err != nil {
		t.Fatalf("Mail nicht lesbar: %v\n%s", err, f.data)
	}
	return msg
}

func testMailer(f *fakeSMTP) *Mailer {
	return &Mailer{
		Host: "127.0.0.1", Port: f.port(), Username: "user", Password: "pw",
		From: "website@example.org", To: []string{"else@example.org"},
		SiteURL: "https://example.org", Insecure: true, Timeout: 5 * time.Second,
	}
}

func body(t *testing.T, m *mail.Message) string {
	b, err := io.ReadAll(quotedprintable.NewReader(m.Body))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestContactMail(t *testing.T) {
	f := startFake(t)
	m := testMailer(f)
	m.NewContactMessage("Jürgen Müller", "juergen@example.com", "Hallo Else,\nkannst du am 12.12. moderieren?")
	msg := f.wait(t)

	dec := new(mime.WordDecoder)
	subject, _ := dec.DecodeHeader(msg.Header.Get("Subject"))
	if subject != "Neue Anfrage von Jürgen Müller" {
		t.Errorf("Betreff = %q", subject)
	}
	reply, err := mail.ParseAddress(msg.Header.Get("Reply-To"))
	if err != nil || reply.Address != "juergen@example.com" {
		t.Errorf("Reply-To = %q", msg.Header.Get("Reply-To"))
	}
	b := body(t, msg)
	if !strings.Contains(b, "kannst du am 12.12. moderieren?") || !strings.Contains(b, "https://example.org/manage/#nachrichten") {
		t.Errorf("Text unvollständig:\n%s", b)
	}
	if !f.authUsed || len(f.rcpt) != 1 {
		t.Errorf("auth=%v rcpt=%v", f.authUsed, f.rcpt)
	}
}

// Zeilenumbrüche im Namen dürfen keine zusätzlichen Header (z. B. Bcc) erzeugen.
func TestNoHeaderInjection(t *testing.T) {
	f := startFake(t)
	m := testMailer(f)
	m.NewContactMessage("Max\r\nBcc: opfer@example.com", "max@example.com", "Test")
	msg := f.wait(t)
	if msg.Header.Get("Bcc") != "" {
		t.Fatalf("Header-Injection möglich: Bcc = %q", msg.Header.Get("Bcc"))
	}
}

func TestOrderMail(t *testing.T) {
	f := startFake(t)
	m := testMailer(f)
	m.NewOrder(Order{Number: "EE-00007", Total: "41,40 €", Lines: []string{"2× Buch (28,00 €)", "Versand 4,90 €"},
		Address: "Erika Mustermann\nMusterstraße 1\n04109 Leipzig", Email: "erika@example.com", Name: "Erika Mustermann"})
	b := body(t, f.wait(t))
	for _, want := range []string{"EE-00007", "41,40 €", "2× Buch", "Musterstraße 1", "/manage/#bestellungen"} {
		if !strings.Contains(b, want) {
			t.Errorf("%q fehlt in:\n%s", want, b)
		}
	}
}

func TestRequiresTLS(t *testing.T) {
	f := startFake(t)
	m := testMailer(f)
	m.Insecure = false
	err := m.Send(context.Background(), "x", "y", "", "")
	if err == nil || !strings.Contains(err.Error(), "STARTTLS") {
		t.Errorf("ohne Verschlüsselung muss der Versand abbrechen, err = %v", err)
	}
}

func TestFromEnv(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

	if m, err := FromEnv(env(nil)); m != nil || err != nil {
		t.Errorf("ohne SMTP_HOST: %v %v", m, err)
	}
	if _, err := FromEnv(env(map[string]string{"SMTP_HOST": "smtp.example.org", "SMTP_FROM": "a@example.org"})); err == nil {
		t.Error("fehlendes NOTIFY_EMAIL muss ein Fehler sein")
	}
	m, err := FromEnv(env(map[string]string{
		"SMTP_HOST": "smtp.example.org", "SMTP_PORT": "465", "SMTP_USERNAME": "website@example.org",
		"NOTIFY_EMAIL": "else@example.org, hendrik@example.org",
	}))
	if err != nil || m.Port != 465 || len(m.To) != 2 || m.From != "website@example.org" || m.SiteURL != "https://elisa-adam.com" {
		t.Errorf("FromEnv: %+v %v", m, err)
	}
}
