package apikey

import (
	"testing"
)

func TestGenerate_uniqueHash(t *testing.T) {
	p1, h1, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	p2, h2, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	if p1 == p2 || h1 == h2 {
		t.Fatal("expected distinct keys")
	}
	if len(h1) != 64 {
		t.Fatalf("hash len %d", len(h1))
	}
}
