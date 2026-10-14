package auth

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestDigestPasswordDeterministic(t *testing.T) {
	salt, _ := hex.DecodeString("00112233445566778899aabbccddeeff")
	first := digestPassword(salt, "secret123")
	second := digestPassword(salt, "secret123")
	if first != second {
		t.Fatalf("digestPassword is not deterministic for the same salt and password")
	}
	other := digestPassword(salt, "secret124")
	if first == other {
		t.Fatalf("digestPassword ignores the password")
	}
}

func TestHashVerifyRoundTrip(t *testing.T) {
	encoded, err := hashPassword("secret123")
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 2 {
		t.Fatalf("hash format should be salt$digest, got %q", encoded)
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("salt is not hex: %v", err)
	}
	if len(salt) != 16 {
		t.Fatalf("salt should be 16 bytes, got %d", len(salt))
	}
	expected := digestPassword(salt, "secret123")
	if hex.EncodeToString(expected[:]) != parts[1] {
		t.Fatalf("hashPassword and digestPassword disagree:\n hashPassword: %s\n digestPassword: %s", parts[1], hex.EncodeToString(expected[:]))
	}
	if !verifyPassword(encoded, "secret123") {
		t.Fatalf("verifyPassword rejected the correct password")
	}
	if verifyPassword(encoded, "secret124") {
		t.Fatalf("verifyPassword accepted a wrong password")
	}
}