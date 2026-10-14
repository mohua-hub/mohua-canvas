package domain

import (
	"net"
	"net/url"
	"strings"
)

type Model struct { Name string; Kind Kind }
type Channel struct {
	ID, Name, PluginID, BaseURL, CredentialRef string
	Models []Model
}
func (c Channel) Validate() error {
	u, err := url.Parse(c.BaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil ||
		u.RawQuery != "" || u.Fragment != "" || len(c.BaseURL) > 2000 ||
		!ValidID(c.ID) || !ValidName(c.Name) || c.PluginID == "" || len(c.PluginID) > 200 ||
		len(c.Models) == 0 || len(c.Models) > 50 ||
		(u.Scheme == "http" && !loopbackHost(u.Hostname())) {
		return Fail("INVALID_INPUT")
	}
	seen := map[string]bool{}
	for _, m := range c.Models {
		key := string(m.Kind)+"/"+m.Name
		if m.Name == "" || len(m.Name) > 200 || !m.Kind.Valid() || seen[key] { return Fail("INVALID_INPUT") }
		seen[key] = true
	}
	return nil
}

func loopbackHost(host string) bool {
	if strings.EqualFold(host,"localhost") { return true }
	ip:=net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
func (c Channel) Supports(model string, kind Kind) bool {
	for _, m := range c.Models { if m.Name == model && m.Kind == kind { return true } }
	return false
}
