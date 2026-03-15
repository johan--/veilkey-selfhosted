package crypto

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"
)

// === MigrateShortHash: 8-char hash format validation ===

func TestGenerateShortHash_Format(t *testing.T) {
	testCases := []struct {
		name  string
		input []byte
	}{
		{"normal ascii", []byte("hello-world")},
		{"empty", []byte("")},
		{"single byte", []byte{0x00}},
		{"all zeros", make([]byte, 32)},
		{"all 0xff", bytes.Repeat([]byte{0xff}, 32)},
		{"binary data", []byte{0xDE, 0xAD, 0xBE, 0xEF}},
		{"unicode korean", []byte("안녕하세요")},
		{"emoji", []byte("🔐🔑")},
		{"very long", bytes.Repeat([]byte("A"), 100000)},
		{"newlines", []byte("line1\nline2\nline3")},
		{"null bytes", []byte("before\x00after")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := GenerateShortHash(tc.input)

			// Must be exactly 11 chars: "CK:" + 8 hex
			if len(hash) != 11 {
				t.Errorf("hash length %d, want 11: %q", len(hash), hash)
			}

			// Must start with "CK:"
			if !strings.HasPrefix(hash, "CK:") {
				t.Errorf("hash must start with CK:, got %q", hash)
			}

			// Hex part must be valid lowercase hex
			hexPart := hash[3:]
			for _, c := range hexPart {
				if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
					t.Errorf("invalid hex char %c in %q", c, hash)
				}
			}
		})
	}
}

func TestGenerateShortHash_Deterministic(t *testing.T) {
	input := []byte("deterministic-test-value")
	h1 := GenerateShortHash(input)
	for i := 0; i < 100; i++ {
		h := GenerateShortHash(input)
		if h != h1 {
			t.Fatalf("iteration %d: hash changed from %q to %q", i, h1, h)
		}
	}
}

func TestGenerateShortHash_Uniqueness(t *testing.T) {
	// Generate 1000 hashes from sequential inputs, check for collisions
	seen := make(map[string]int)
	n := 1000
	for i := 0; i < n; i++ {
		input := []byte(fmt.Sprintf("credential-value-%d", i))
		hash := GenerateShortHash(input)
		if prev, exists := seen[hash]; exists {
			t.Errorf("collision: input %d and %d both produce %s", prev, i, hash)
		}
		seen[hash] = i
	}
	t.Logf("Generated %d unique hashes from %d inputs (0 collisions)", len(seen), n)
}

func TestGenerateShortHash_ConcurrentSafety(t *testing.T) {
	var wg sync.WaitGroup
	results := make([]string, 100)
	input := []byte("concurrent-test-value")

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = GenerateShortHash(input)
		}(i)
	}
	wg.Wait()

	// All results should be identical
	for i, r := range results {
		if r != results[0] {
			t.Errorf("goroutine %d got %q, expected %q", i, r, results[0])
		}
	}
}

// === Unicode / special character encryption ===

func TestEncryptDecrypt_Unicode(t *testing.T) {
	key, _ := GenerateKey()

	testCases := []struct {
		name  string
		value string
	}{
		{"korean", "비밀번호123!@#"},
		{"japanese", "パスワード"},
		{"chinese", "密码测试"},
		{"emoji", "🔐secret🔑key"},
		{"mixed", "admin@서버.com:Pässwörd!"},
		{"rtl arabic", "كلمة المرور"},
		{"combining chars", "e\u0301e\u0301"}, // é with combining accent
		{"zero width", "pass\u200Bword"},       // zero-width space
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plaintext := []byte(tc.value)

			ct, nonce, err := Encrypt(key, plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			pt, err := Decrypt(key, ct, nonce)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if !bytes.Equal(pt, plaintext) {
				t.Errorf("round-trip failed: got %q, want %q", pt, plaintext)
			}

			if !utf8.Valid(pt) && utf8.Valid(plaintext) {
				t.Error("UTF-8 validity lost after round-trip")
			}
		})
	}
}

func TestEncryptDecrypt_SpecialValues(t *testing.T) {
	key, _ := GenerateKey()

	testCases := []struct {
		name  string
		value []byte
	}{
		{"single null", []byte{0}},
		{"null in middle", []byte("before\x00after")},
		{"all nulls 32B", make([]byte, 32)},
		{"1 byte", []byte{0x42}},
		{"max single byte", []byte{0xFF}},
		{"1MB", bytes.Repeat([]byte("X"), 1<<20)},
		{"binary sequence", func() []byte {
			b := make([]byte, 256)
			for i := range b {
				b[i] = byte(i)
			}
			return b
		}()},
		{"json payload", []byte(`{"password":"s3cret","token":"abc123"}`)},
		{"multiline yaml", []byte("key1: value1\nkey2: value2\nlist:\n  - item1\n  - item2\n")},
		{"shell dangerous", []byte(`$(rm -rf /); echo "pwned"`)},
		{"sql injection", []byte(`'; DROP TABLE credentials; --`)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ct, nonce, err := Encrypt(key, tc.value)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			pt, err := Decrypt(key, ct, nonce)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if !bytes.Equal(pt, tc.value) {
				t.Errorf("round-trip failed for %q", tc.name)
			}
		})
	}
}

// === Nonce uniqueness under concurrent generation ===

func TestNonceUniqueness_Concurrent(t *testing.T) {
	var mu sync.Mutex
	nonces := make(map[string]bool)
	var wg sync.WaitGroup
	collisions := 0

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			nonce, err := GenerateNonce()
			if err != nil {
				t.Errorf("GenerateNonce failed: %v", err)
				return
			}
			key := string(nonce)
			mu.Lock()
			if nonces[key] {
				collisions++
			}
			nonces[key] = true
			mu.Unlock()
		}()
	}
	wg.Wait()

	if collisions > 0 {
		t.Errorf("nonce collisions detected: %d out of 1000", collisions)
	}
}

// === Semantic security: same plaintext → different ciphertext ===

func TestSemanticSecurity_DifferentNonces(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := []byte("same-secret-value")

	ciphertexts := make([][]byte, 50)
	for i := 0; i < 50; i++ {
		ct, _, err := Encrypt(key, plaintext)
		if err != nil {
			t.Fatalf("Encrypt #%d failed: %v", i, err)
		}
		ciphertexts[i] = ct
	}

	// All ciphertexts should be different
	for i := 0; i < len(ciphertexts); i++ {
		for j := i + 1; j < len(ciphertexts); j++ {
			if bytes.Equal(ciphertexts[i], ciphertexts[j]) {
				t.Errorf("identical ciphertext at indices %d and %d (nonce reuse?)", i, j)
			}
		}
	}
}

// === GCM ciphertext size validation ===

func TestGCM_CiphertextSize(t *testing.T) {
	key, _ := GenerateKey()

	// GCM adds 16-byte authentication tag
	sizes := []int{0, 1, 16, 100, 1000, 10000}
	for _, size := range sizes {
		plaintext := bytes.Repeat([]byte("A"), size)
		ct, _, err := Encrypt(key, plaintext)
		if err != nil {
			t.Fatalf("Encrypt failed for size %d: %v", size, err)
		}
		expectedCTSize := size + 16 // plaintext + GCM tag
		if len(ct) != expectedCTSize {
			t.Errorf("size %d: ciphertext %d bytes, expected %d", size, len(ct), expectedCTSize)
		}
	}
}

// === Truncated/corrupted ciphertext handling ===

func TestDecrypt_CorruptedInput(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := []byte("test-value")
	ct, nonce, _ := Encrypt(key, plaintext)

	testCases := []struct {
		name       string
		ciphertext []byte
		nonce      []byte
	}{
		{"empty ciphertext", []byte{}, nonce},
		{"truncated ciphertext", ct[:len(ct)/2], nonce},
		{"one byte ciphertext", ct[:1], nonce},
		// Note: wrong nonce length causes GCM panic (by design), not tested here
		{"flipped bit in tag", func() []byte {
			c := make([]byte, len(ct))
			copy(c, ct)
			c[len(c)-1] ^= 0x01 // flip last bit of GCM tag
			return c
		}(), nonce},
		{"flipped bit in body", func() []byte {
			c := make([]byte, len(ct))
			copy(c, ct)
			c[0] ^= 0x01
			return c
		}(), nonce},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Decrypt(key, tc.ciphertext, tc.nonce)
			if err == nil {
				t.Error("expected decryption to fail for corrupted input")
			}
		})
	}
}
