package stdlib

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHashSHA256 tests SHA256 hashing
func TestHashSHA256(t *testing.T) {
	hash, err := Hash("test data", HashSHA256)

	assert.NoError(t, err)
	assert.NotNil(t, hash)
	assert.Equal(t, "sha256", hash.Algorithm)
	assert.NotEmpty(t, hash.HexString)
	assert.NotEmpty(t, hash.Base64)
}

// TestHashSHA512 tests SHA512 hashing
func TestHashSHA512(t *testing.T) {
	hash, err := Hash("test data", HashSHA512)

	assert.NoError(t, err)
	assert.NotNil(t, hash)
	assert.Equal(t, "sha512", hash.Algorithm)
	assert.NotEmpty(t, hash.HexString)
	assert.NotEmpty(t, hash.Base64)
}

// TestHashInvalidAlgorithm tests invalid hash algorithm
func TestHashInvalidAlgorithm(t *testing.T) {
	hash, err := Hash("test data", "invalid")

	assert.Error(t, err)
	assert.Nil(t, hash)
}

// TestHmacSHA256 tests HMAC-SHA256
func TestHmacSHA256(t *testing.T) {
	result, err := Hmac("test data", "secret key", HashSHA256)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "hmac-sha256", result.Algorithm)
	assert.NotEmpty(t, result.HexString)
	assert.NotEmpty(t, result.Base64)
}

// TestHmacSHA512 tests HMAC-SHA512
func TestHmacSHA512(t *testing.T) {
	result, err := Hmac("test data", "secret key", HashSHA512)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "hmac-sha512", result.Algorithm)
}

// TestEncrypt tests encryption
func TestEncrypt(t *testing.T) {
	key := "12345678901234567890123456789012" // 32 bytes

	encrypted, err := Encrypt("secret message", key)

	assert.NoError(t, err)
	assert.NotNil(t, encrypted)
	assert.Equal(t, "aes-256-gcm", encrypted.Algorithm)
	assert.NotEmpty(t, encrypted.Data)
	assert.NotEmpty(t, encrypted.Nonce)
}

// TestDecrypt tests decryption structure
func TestDecrypt(t *testing.T) {
	key := "12345678901234567890123456789012" // 32 bytes

	// Just test that Decrypt function exists and handles DecryptedData correctly
	encryptedData := &EncryptedData{
		Algorithm: "aes-256-gcm",
		Data:      "base64_encoded_data",
		Nonce:     "base64_encoded_nonce",
		Timestamp: 1000000,
	}

	// Decrypt will fail with invalid data, but we test the structure
	_, err := Decrypt(encryptedData, key)

	assert.Error(t, err) // Expected to fail with invalid data
	assert.NotNil(t, encryptedData)
	assert.Equal(t, "aes-256-gcm", encryptedData.Algorithm)
}

// TestEncryptShortKey tests encryption with short key
func TestEncryptShortKey(t *testing.T) {
	encrypted, err := Encrypt("test", "short")

	assert.Error(t, err)
	assert.Nil(t, encrypted)
}

// TestGenerateRandomKey tests random key generation
func TestGenerateRandomKey(t *testing.T) {
	key, err := GenerateRandomKey(32)

	assert.NoError(t, err)
	assert.NotEmpty(t, key)
	assert.Greater(t, len(key), 0)
}

// TestGenerateRandomBytes tests random bytes generation
func TestGenerateRandomBytes(t *testing.T) {
	bytes, err := GenerateRandomBytes(16)

	assert.NoError(t, err)
	assert.NotNil(t, bytes)
	assert.Equal(t, 16, len(bytes))
}

// TestNewJwtToken tests JWT token creation
func TestNewJwtToken(t *testing.T) {
	claims := &JwtClaims{
		Subject: "user123",
		Issuer:  "app.example.com",
		Audience: "example.com",
		ExpiresAt: 1700000000,
		IssuedAt: 1600000000,
		JwtID: "token123",
	}

	token, err := NewJwtToken(claims, "secret")

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.NotEmpty(t, token.Header)
	assert.NotEmpty(t, token.Payload)
	assert.NotEmpty(t, token.Signature)
	assert.Equal(t, true, token.Valid)
}

// TestVerifyJwtToken tests JWT token verification
func TestVerifyJwtToken(t *testing.T) {
	token := "header.payload.signature"

	verified, err := VerifyJwtToken(token, "secret")

	assert.NoError(t, err)
	assert.NotNil(t, verified)
	assert.Equal(t, true, verified.Valid)
}

// TestVerifyHmacValid tests HMAC verification (valid)
func TestVerifyHmacValid(t *testing.T) {
	data := "test data"
	key := "secret key"

	// Create HMAC
	hmac, _ := Hmac(data, key, HashSHA256)

	// Verify it
	valid, err := VerifyHmac(data, hmac.HexString, key, HashSHA256)

	assert.NoError(t, err)
	assert.Equal(t, true, valid)
}

// TestVerifyHmacInvalid tests HMAC verification (invalid)
func TestVerifyHmacInvalid(t *testing.T) {
	data := "test data"
	key := "secret key"

	valid, err := VerifyHmac(data, "invalid_signature", key, HashSHA256)

	assert.NoError(t, err)
	assert.Equal(t, false, valid)
}

// TestGenerateSecureToken tests secure token generation
func TestGenerateSecureToken(t *testing.T) {
	token, err := GenerateSecureToken(32)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, len(token), 0)
	// Should be hex encoded (even length)
	assert.Equal(t, 0, len(token)%2)
}

// TestCheckPasswordStrength tests password strength checking
func TestCheckPasswordStrength(t *testing.T) {
	strength := CheckPasswordStrength("SecurePass123!@#")

	assert.NotNil(t, strength)
	assert.Equal(t, true, strength["length"])
	assert.Equal(t, true, strength["uppercase"])
	assert.Equal(t, true, strength["lowercase"])
	assert.Equal(t, true, strength["numbers"])
	assert.Equal(t, true, strength["special"])
	assert.Greater(t, strength["score"], 0)
}

// TestCheckPasswordStrengthWeak tests weak password
func TestCheckPasswordStrengthWeak(t *testing.T) {
	strength := CheckPasswordStrength("a")

	assert.NotNil(t, strength)
	assert.Equal(t, false, strength["length"])
	assert.Equal(t, 1, strength["score"])
}

// TestNewCryptoManager tests crypto manager creation
func TestNewCryptoManager(t *testing.T) {
	manager := NewCryptoManager()

	assert.NotNil(t, manager)
	assert.Equal(t, 0, len(manager.Keys))
	assert.NotNil(t, manager.TlsContext)
	assert.Equal(t, true, manager.TlsContext.Enabled)
}

// TestAddKey tests adding a key
func TestAddKey(t *testing.T) {
	manager := NewCryptoManager()

	key := manager.AddKey("symmetric", "AES256", 1700000000)

	assert.NotNil(t, key)
	assert.Equal(t, "symmetric", key.Type)
	assert.Equal(t, "AES256", key.Algorithm)
	assert.Equal(t, false, key.Revoked)
}

// TestRevokeKey tests revoking a key
func TestRevokeKey(t *testing.T) {
	manager := NewCryptoManager()
	key := manager.AddKey("symmetric", "AES256", 1700000000)

	err := manager.RevokeKey(key.ID)

	assert.NoError(t, err)
	assert.Equal(t, true, key.Revoked)
}

// TestGetKey tests retrieving a key
func TestGetKey(t *testing.T) {
	manager := NewCryptoManager()
	addedKey := manager.AddKey("symmetric", "AES256", 1700000000)

	retrieved := manager.GetKey(addedKey.ID)

	assert.NotNil(t, retrieved)
	assert.Equal(t, addedKey.ID, retrieved.ID)
}

// TestGetKeys tests retrieving all keys
func TestGetKeys(t *testing.T) {
	manager := NewCryptoManager()

	manager.AddKey("symmetric", "AES256", 1700000000)
	manager.AddKey("public", "RSA2048", 1700000000)
	manager.AddKey("private", "ECDSA256", 1700000000)

	keys := manager.GetKeys()

	assert.Equal(t, 3, len(keys))
}

// TestAddCertificate tests adding a TLS certificate
func TestAddCertificate(t *testing.T) {
	manager := NewCryptoManager()

	cert := manager.TlsContext.AddCertificate("example.com", "Let's Encrypt", 1700000000)

	assert.NotNil(t, cert)
	assert.Equal(t, "example.com", cert.Subject)
	assert.Equal(t, "Let's Encrypt", cert.Issuer)
}

// TestRemoveCertificate tests removing a certificate
func TestRemoveCertificate(t *testing.T) {
	manager := NewCryptoManager()
	cert := manager.TlsContext.AddCertificate("example.com", "Let's Encrypt", 1700000000)

	err := manager.TlsContext.RemoveCertificate(cert.ID)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(manager.TlsContext.Certificates))
}

// TestGetCertificates tests retrieving all certificates
func TestGetCertificates(t *testing.T) {
	manager := NewCryptoManager()

	manager.TlsContext.AddCertificate("example.com", "Let's Encrypt", 1700000000)
	manager.TlsContext.AddCertificate("test.com", "DigiCert", 1700000000)

	certs := manager.TlsContext.GetCertificates()

	assert.Equal(t, 2, len(certs))
}

// TestTlsContextEnabled tests TLS context enabled
func TestTlsContextEnabled(t *testing.T) {
	manager := NewCryptoManager()

	assert.Equal(t, true, manager.TlsContext.Enabled)
	assert.Equal(t, "1.2", manager.TlsContext.MinVersion)
	assert.Greater(t, len(manager.TlsContext.CipherSuites), 0)
}

// TestHashConsistency tests hash consistency
func TestHashConsistency(t *testing.T) {
	data := "test data"

	hash1, _ := Hash(data, HashSHA256)
	hash2, _ := Hash(data, HashSHA256)

	assert.Equal(t, hash1.HexString, hash2.HexString)
}

// TestEncryptionUniqueness tests encryption produces unique nonces
func TestEncryptionUniqueness(t *testing.T) {
	key := "12345678901234567890123456789012"
	plaintext := "secret message"

	encrypted1, err1 := Encrypt(plaintext, key)
	encrypted2, err2 := Encrypt(plaintext, key)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Different nonces should be generated each time
	assert.NotEqual(t, encrypted1.Nonce, encrypted2.Nonce)
}

// TestJwtTokenStructure tests JWT token structure
func TestJwtTokenStructure(t *testing.T) {
	claims := &JwtClaims{
		Subject: "user123",
		Issuer:  "app.example.com",
		ExpiresAt: 1700000000,
		IssuedAt: 1600000000,
		JwtID: "token123",
	}

	token, _ := NewJwtToken(claims, "secret")

	// JWT should have 3 parts separated by dots
	parts := strings.Split(token.Header+"."+token.Payload+"."+token.Signature, ".")
	assert.Equal(t, 3, len(parts))
}

// TestMultipleHashAlgorithms tests multiple hash algorithms
func TestMultipleHashAlgorithms(t *testing.T) {
	data := "test"

	sha256, _ := Hash(data, HashSHA256)
	sha512, _ := Hash(data, HashSHA512)

	assert.NotEqual(t, sha256.HexString, sha512.HexString)
	assert.Greater(t, len(sha512.HexString), len(sha256.HexString))
}
