package user

import "testing"

func TestRootDomainFromHost(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{"www.niceboxs.com", "niceboxs.com"},
		{"WWW.Niceboxs.COM", "niceboxs.com"},
		{"www.niceboxs.com:443", "niceboxs.com"},
		{"niceboxs.com", "niceboxs.com"},
		{"a.b.example.co.uk", "co.uk"},
	}
	for _, tc := range tests {
		got := rootDomainFromHost(tc.host)
		if got != tc.want {
			t.Errorf("rootDomainFromHost(%q) = %q, want %q", tc.host, got, tc.want)
		}
	}
}
