// Package policy loads machine-administered browser access policy.
package policy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Policy struct {
	DefaultAction  string   `json:"defaultAction"`
	AllowedOrigins []string `json:"allowedOrigins"`
	BlockedOrigins []string `json:"blockedOrigins"`
}

func Default() Policy { return Policy{"ask", []string{}, []string{}} }
func Parse(data []byte) (Policy, error) {
	if len(bytes.TrimSpace(data)) == 0 || bytes.TrimSpace(data)[0] != '{' {
		return Policy{}, fmt.Errorf("policy must be an object")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return Policy{}, err
	}
	for name, value := range fields {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return Policy{}, fmt.Errorf("policy field %s cannot be null", name)
		}
	}
	p := Default()
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&p); err != nil {
		return p, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return p, fmt.Errorf("unexpected trailing policy data")
	}
	if p.DefaultAction != "ask" && p.DefaultAction != "allow" && p.DefaultAction != "deny" {
		return p, fmt.Errorf("invalid defaultAction")
	}
	for _, list := range [][]string{p.AllowedOrigins, p.BlockedOrigins} {
		for _, origin := range list {
			if _, err := Origin(origin); err != nil {
				return p, err
			}
		}
	}
	return p, nil
}
func Origin(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") || strings.ContainsAny(raw, "*\\\r\n\t ") {
		return "", fmt.Errorf("expected an exact HTTPS origin")
	}
	host := strings.ToLower(u.Host)
	if u.Port() == "443" {
		host = strings.TrimSuffix(host, ":443")
	}
	return "https://" + host, nil
}
func (p Policy) Action(origin string) (string, error) {
	origin, err := Origin(origin)
	if err != nil {
		return "deny", err
	}
	for _, v := range p.BlockedOrigins {
		normalized, _ := Origin(v)
		if normalized == origin {
			return "deny", nil
		}
	}
	for _, v := range p.AllowedOrigins {
		normalized, _ := Origin(v)
		if normalized == origin {
			return "allow", nil
		}
	}
	return p.DefaultAction, nil
}
func Load() (Policy, error) {
	data, found, err := machinePolicy()
	if err != nil {
		return Policy{}, err
	}
	if found {
		return Parse(data)
	}
	root, err := machineDataDir()
	if err != nil {
		return Policy{}, err
	}
	data, err = os.ReadFile(filepath.Join(root, "eid-pd-reader", "policy.json"))
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return Policy{}, err
	}
	return Parse(data)
}

func DataDirectory() (string, error) { return machineDataDir() }
