package crypto

import (
	"bytes"
	"testing"
)

// TestDeriveKEK verifies that KEK derivation is deterministic and that
// different inputs produce different keys.
func TestDeriveKEK(t *testing.T) {
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt: %v", err)
	}

	password := "test-password-for-kek-derivation"

	// Derive KEK twice with same password + salt
	kek1 := DeriveKEK(password, salt)
	kek2 := DeriveKEK(password, salt)

	// Must be deterministic
	if !bytes.Equal(kek1, kek2) {
		t.Error("DeriveKEK: same password+salt should produce identical KEK")
	}

	// Must be the declared KEK size
	if len(kek1) != KEKSize {
		t.Errorf("DeriveKEK: got %d bytes, want %d (KEKSize)", len(kek1), KEKSize)
	}

	// Different password → different KEK
	kekWrongPass := DeriveKEK("wrong-password", salt)
	if bytes.Equal(kek1, kekWrongPass) {
		t.Error("DeriveKEK: different password should produce different KEK")
	}

	// Different salt → different KEK (same password)
	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt (second): %v", err)
	}
	kekOtherSalt := DeriveKEK(password, salt2)
	if bytes.Equal(kek1, kekOtherSalt) {
		t.Error("DeriveKEK: different salt should produce different KEK")
	}

	// Empty password still derives a key (no panic)
	kekEmpty := DeriveKEK("", salt)
	if len(kekEmpty) != KEKSize {
		t.Errorf("DeriveKEK with empty password: got %d bytes, want %d", len(kekEmpty), KEKSize)
	}
	if bytes.Equal(kekEmpty, kek1) {
		t.Error("DeriveKEK: empty password should differ from non-empty password")
	}

	// Unicode password round-trip
	unicodePass := "P@$$w0rd-日本語-émojis🔐"
	kekUni1 := DeriveKEK(unicodePass, salt)
	kekUni2 := DeriveKEK(unicodePass, salt)
	if !bytes.Equal(kekUni1, kekUni2) {
		t.Error("DeriveKEK: unicode password should be deterministic")
	}
	if len(kekUni1) != KEKSize {
		t.Errorf("DeriveKEK unicode: got %d bytes, want %d", len(kekUni1), KEKSize)
	}
}

// TestDeriveKEK_Usable confirms that a KEK derived from DeriveKEK can
// actually encrypt and decrypt a DEK (end-to-end usability check).
func TestDeriveKEK_Usable(t *testing.T) {
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt: %v", err)
	}

	kek := DeriveKEK("usable-password", salt)

	// Generate a random DEK
	dek, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	// Encrypt DEK with derived KEK
	encDEK, nonce, err := EncryptDEK(kek, dek)
	if err != nil {
		t.Fatalf("EncryptDEK: %v", err)
	}

	// Derive KEK again (simulates server restart)
	kek2 := DeriveKEK("usable-password", salt)

	// Should be able to decrypt DEK with re-derived KEK
	decDEK, err := DecryptDEK(kek2, encDEK, nonce)
	if err != nil {
		t.Fatalf("DecryptDEK with re-derived KEK: %v", err)
	}

	if !bytes.Equal(decDEK, dek) {
		t.Error("DecryptDEK: re-derived KEK should produce the original DEK")
	}

	// Wrong password → KEK derivation produces wrong key → decryption fails
	wrongKEK := DeriveKEK("wrong-password", salt)
	_, err = DecryptDEK(wrongKEK, encDEK, nonce)
	if err == nil {
		t.Error("DecryptDEK with wrong KEK should fail, but succeeded")
	}
}

// TestGenerateSalt verifies that GenerateSalt produces the correct size and
// that repeated calls produce unique values.
func TestGenerateSalt(t *testing.T) {
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt: %v", err)
	}

	// Must be exactly SaltSize bytes
	if len(salt) != SaltSize {
		t.Errorf("GenerateSalt: got %d bytes, want %d (SaltSize)", len(salt), SaltSize)
	}

	// Salt must not be all zeros (astronomically unlikely with a proper RNG)
	allZero := true
	for _, b := range salt {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("GenerateSalt: produced all-zero salt (broken RNG?)")
	}

	// Two salts must be different (collision probability is negligible)
	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt (second): %v", err)
	}
	if bytes.Equal(salt, salt2) {
		t.Error("GenerateSalt: two calls should produce different salts")
	}

	// Third call also unique
	salt3, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt (third): %v", err)
	}
	if bytes.Equal(salt, salt3) || bytes.Equal(salt2, salt3) {
		t.Error("GenerateSalt: salts should be unique across calls")
	}
}

// TestGenerateSalt_Size confirms the SaltSize constant is reasonable.
func TestGenerateSalt_Size(t *testing.T) {
	// SaltSize must be at least 16 bytes for security
	if SaltSize < 16 {
		t.Errorf("SaltSize = %d, must be >= 16 for security", SaltSize)
	}
}

// TestKDFConstants validates that the KDF constants meet security minimums.
func TestKDFConstants(t *testing.T) {
	// NIST recommends >= 600,000 iterations for PBKDF2-SHA256 as of 2023
	if KDFIterations < 100000 {
		t.Errorf("KDFIterations = %d, should be >= 100000 for security", KDFIterations)
	}

	// KEKSize must be 32 bytes for AES-256
	if KEKSize != 32 {
		t.Errorf("KEKSize = %d, expected 32 (AES-256)", KEKSize)
	}
}
