package auth

import (
	"testing"
	"time"
)

func TestIssueAndParseUserToken(t *testing.T) {
	secret := "test-secret"
	token, err := IssueUserToken(42, secret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	uid, err := ParseUserToken(token, secret)
	if err != nil {
		t.Fatal(err)
	}
	if uid != 42 {
		t.Fatalf("want uid 42, got %d", uid)
	}
}
