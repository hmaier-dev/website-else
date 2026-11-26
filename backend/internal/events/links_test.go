package events

import "testing"

func TestSplitLink(t *testing.T) {
	for in, want := range map[string][2]string{
		"https://example.org/x":             {"https://example.org/x", ""},
		"  https://kolinar.net":             {"https://kolinar.net", ""},
		"null":                              {"", ""},
		"NULL":                              {"", ""},
		"":                                  {"", ""},
		"Eintritt: Spendenbasis":            {"", "Eintritt: Spendenbasis"},
		"Anmeldung per Mail oder Instagram": {"", "Anmeldung per Mail oder Instagram"},
		"javascript:alert(1)":               {"", ""},
		"mailto:else@example.org":           {"", ""},
	} {
		l, n := splitLink(in)
		if l != want[0] || n != want[1] {
			t.Errorf("splitLink(%q) = %q, %q; erwartet %q, %q", in, l, n, want[0], want[1])
		}
	}
	if validLinkInput("javascript:alert(1)") == "" {
		t.Error("javascript:-Link wurde akzeptiert")
	}
	if validLinkInput("Eintritt: Spendenbasis") != "" {
		t.Error("Hinweistext wurde abgelehnt")
	}
}
