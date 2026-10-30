package policy

import "testing"

func TestRules(t *testing.T) {
	p, err := Parse([]byte(`{"defaultAction":"ask","allowedOrigins":["https://a.test:443"],"blockedOrigins":["https://b.test"]}`))
	if err != nil {
		t.Fatal(err)
	}
	for origin, want := range map[string]string{"https://a.test": "allow", "https://b.test": "deny", "https://c.test": "ask"} {
		got, err := p.Action(origin)
		if err != nil || got != want {
			t.Fatalf("%s: %s %v", origin, got, err)
		}
	}
	p.DefaultAction = "allow"
	p.AllowedOrigins = append(p.AllowedOrigins, "https://b.test")
	if got, _ := p.Action("https://b.test"); got != "deny" {
		t.Fatal("deny must win")
	}
	p.DefaultAction = "deny"
	if got, _ := p.Action("https://c.test"); got != "deny" {
		t.Fatal("closed mode")
	}
}
func TestMalformed(t *testing.T) {
	for _, s := range []string{`null null`, `{"defaultAction":null}`, `{"allowedOrigins":null}`, `{"blockedOrigins":null}`, `{"defaultAction":"oops"}`, `{"unknown":true}`, `{"allowedOrigins":["http://localhost"]}`, `{"allowedOrigins":["https://*.test"]}`, `{"blockedOrigins":["https://user@host"]}`, `{"allowedOrigins":["https://a.test/path"]}`} {
		if _, err := Parse([]byte(s)); err == nil {
			t.Fatalf("accepted %s", s)
		}
	}
}
