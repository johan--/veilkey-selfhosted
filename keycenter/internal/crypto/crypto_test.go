package crypto

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(key))
	}

	// Two keys should be different
	key2, _ := GenerateKey()
	if bytes.Equal(key, key2) {
		t.Error("two generated keys should be different")
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce failed: %v", err)
	}
	if len(nonce) != 12 {
		t.Errorf("expected 12 bytes, got %d", len(nonce))
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := []byte("my-secret-password-123")

	ciphertext, nonce, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Error("ciphertext should differ from plaintext")
	}

	decrypted, err := Decrypt(key, ciphertext, nonce)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted %q != plaintext %q", decrypted, plaintext)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()
	plaintext := []byte("secret")

	ciphertext, nonce, _ := Encrypt(key1, plaintext)

	_, err := Decrypt(key2, ciphertext, nonce)
	if err == nil {
		t.Error("decrypting with wrong key should fail")
	}
}

func TestDecryptTamperedCiphertext(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := []byte("secret")

	ciphertext, nonce, _ := Encrypt(key, plaintext)

	// Tamper with ciphertext
	ciphertext[0] ^= 0xff

	_, err := Decrypt(key, ciphertext, nonce)
	if err == nil {
		t.Error("decrypting tampered ciphertext should fail")
	}
}

func TestEncryptDEKDecryptDEK(t *testing.T) {
	kek, _ := GenerateKey()
	dek, _ := GenerateKey()

	encryptedDEK, nonce, err := EncryptDEK(kek, dek)
	if err != nil {
		t.Fatalf("EncryptDEK failed: %v", err)
	}

	decryptedDEK, err := DecryptDEK(kek, encryptedDEK, nonce)
	if err != nil {
		t.Fatalf("DecryptDEK failed: %v", err)
	}

	if !bytes.Equal(decryptedDEK, dek) {
		t.Error("decrypted DEK should match original")
	}
}

func TestGenerateShortHash(t *testing.T) {
	data := []byte("test-encrypted-value")
	hash := GenerateShortHash(data)

	if !strings.HasPrefix(hash, "CK:") {
		t.Errorf("hash should start with CK:, got %q", hash)
	}

	if len(hash) != 11 { // "CK:" + 8 hex chars
		t.Errorf("expected length 11, got %d (%q)", len(hash), hash)
	}

	// Same input should produce same hash
	hash2 := GenerateShortHash(data)
	if hash != hash2 {
		t.Errorf("same input should produce same hash: %q != %q", hash, hash2)
	}

	// Different input should produce different hash (with high probability)
	hash3 := GenerateShortHash([]byte("different-data"))
	if hash == hash3 {
		t.Error("different input should produce different hash")
	}
}

func TestEncryptEmptyPlaintext(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := []byte("")

	ciphertext, nonce, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt empty plaintext failed: %v", err)
	}

	decrypted, err := Decrypt(key, ciphertext, nonce)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("expected empty, got %q", decrypted)
	}
}

func TestEncryptLargePlaintext(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := bytes.Repeat([]byte("A"), 10000)

	ciphertext, nonce, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt large plaintext failed: %v", err)
	}

	decrypted, err := Decrypt(key, ciphertext, nonce)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("decrypted large plaintext should match original")
	}
}
