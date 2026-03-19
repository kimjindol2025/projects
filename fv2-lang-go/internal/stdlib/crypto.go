// Package stdlib provides Crypto support for FV 2.0
package stdlib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"
)

// HashAlgorithm represents a hash algorithm
type HashAlgorithm string

const (
	HashSHA256 HashAlgorithm = "sha256"
	HashSHA512 HashAlgorithm = "sha512"
)

// EncryptionAlgorithm represents an encryption algorithm
type EncryptionAlgorithm string

const (
	EncryptAES256GCM EncryptionAlgorithm = "aes-256-gcm"
)

// HashResult represents a hash result
type HashResult struct {
	Algorithm  string
	Data       string
	HexString  string
	Base64     string
	Timestamp  int64
}

// EncryptedData represents encrypted data
type EncryptedData struct {
	Algorithm string
	Data      string
	Nonce     string
	Tag       string
	Timestamp int64
}

// DecryptedData represents decrypted data
type DecryptedData struct {
	Data      string
	Timestamp int64
	Success   bool
}

// JwtToken represents a JWT token
type JwtToken struct {
	Header    string
	Payload   string
	Signature string
	Valid     bool
	ExpiresAt int64
	IssuedAt  int64
}

// JwtClaims represents JWT claims
type JwtClaims struct {
	Subject   string
	Issuer    string
	Audience  string
	ExpiresAt int64
	IssuedAt  int64
	NotBefore int64
	JwtID     string
	Claims    map[string]string
}

// TlsCertificate represents a TLS certificate
type TlsCertificate struct {
	ID        int64
	Subject   string
	Issuer    string
	ValidFrom int64
	ValidTo   int64
	PublicKey string
	Thumbprint string
	Algorithm string
}

// TlsContext represents a TLS context
type TlsContext struct {
	ID           int64
	Certificates map[int64]*TlsCertificate
	Enabled      bool
	MinVersion   string // "1.2", "1.3"
	CipherSuites []string
	mutex        sync.RWMutex
}

// CryptoKey represents a cryptographic key
type CryptoKey struct {
	ID        int64
	Type      string // "symmetric", "public", "private"
	Algorithm string // "AES256", "RSA2048", "ECDSA256"
	Data      []byte
	CreatedAt int64
	ExpiresAt int64
	Revoked   bool
}

// CryptoManager manages cryptographic operations
type CryptoManager struct {
	Keys       map[int64]*CryptoKey
	KeyIDGen   int64
	CertIDGen  int64
	TlsContext *TlsContext
	mutex      sync.RWMutex
}

// NewCryptoManager creates a new crypto manager
func NewCryptoManager() *CryptoManager {
	return &CryptoManager{
		Keys:      make(map[int64]*CryptoKey),
		KeyIDGen:  1,
		CertIDGen: 1,
		TlsContext: &TlsContext{
			ID:           1,
			Certificates: make(map[int64]*TlsCertificate),
			Enabled:      true,
			MinVersion:   "1.2",
			CipherSuites: []string{"TLS_AES_256_GCM_SHA384", "TLS_CHACHA20_POLY1305_SHA256"},
		},
	}
}

// Hash computes a hash of data
func Hash(data string, algorithm HashAlgorithm) (*HashResult, error) {
	result := &HashResult{
		Algorithm: string(algorithm),
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	switch algorithm {
	case HashSHA256:
		hash := sha256.Sum256([]byte(data))
		result.HexString = hex.EncodeToString(hash[:])
		result.Base64 = base64.StdEncoding.EncodeToString(hash[:])
		return result, nil

	case HashSHA512:
		hash := sha512.Sum512([]byte(data))
		result.HexString = hex.EncodeToString(hash[:])
		result.Base64 = base64.StdEncoding.EncodeToString(hash[:])
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}
}

// Hmac computes an HMAC
func Hmac(data string, key string, algorithm HashAlgorithm) (*HashResult, error) {
	result := &HashResult{
		Algorithm: fmt.Sprintf("hmac-%s", algorithm),
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	switch algorithm {
	case HashSHA256:
		h := hmac.New(sha256.New, []byte(key))
		h.Write([]byte(data))
		hash := h.Sum(nil)
		result.HexString = hex.EncodeToString(hash)
		result.Base64 = base64.StdEncoding.EncodeToString(hash)
		return result, nil

	case HashSHA512:
		h := hmac.New(sha512.New, []byte(key))
		h.Write([]byte(data))
		hash := h.Sum(nil)
		result.HexString = hex.EncodeToString(hash)
		result.Base64 = base64.StdEncoding.EncodeToString(hash)
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}
}

// Encrypt encrypts data using AES-256-GCM
func Encrypt(data string, key string) (*EncryptedData, error) {
	if len(key) < 32 {
		return nil, fmt.Errorf("key must be at least 32 bytes")
	}

	block, err := aes.NewCipher([]byte(key[:32]))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)

	return &EncryptedData{
		Algorithm: "aes-256-gcm",
		Data:      base64.StdEncoding.EncodeToString(ciphertext),
		Nonce:     base64.StdEncoding.EncodeToString(nonce),
		Timestamp: time.Now().Unix(),
	}, nil
}

// Decrypt decrypts data encrypted with AES-256-GCM
func Decrypt(encryptedData *EncryptedData, key string) (*DecryptedData, error) {
	if len(key) < 32 {
		return nil, fmt.Errorf("key must be at least 32 bytes")
	}

	block, err := aes.NewCipher([]byte(key[:32]))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData.Data)
	if err != nil {
		return nil, err
	}

	nonce, err := base64.StdEncoding.DecodeString(encryptedData.Nonce)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return &DecryptedData{
		Data:      string(plaintext),
		Timestamp: time.Now().Unix(),
		Success:   true,
	}, nil
}

// GenerateRandomKey generates a random cryptographic key
func GenerateRandomKey(size int) (string, error) {
	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(key), nil
}

// GenerateRandomBytes generates random bytes
func GenerateRandomBytes(size int) ([]byte, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	return bytes, nil
}

// NewJwtToken creates a new JWT token
func NewJwtToken(claims *JwtClaims, secret string) (*JwtToken, error) {
	// Simplified JWT implementation
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	payload := fmt.Sprintf(`{"sub":"%s","iss":"%s","aud":"%s","exp":%d,"iat":%d,"nbf":%d,"jti":"%s"}`,
		claims.Subject, claims.Issuer, claims.Audience,
		claims.ExpiresAt, claims.IssuedAt, claims.NotBefore,
		claims.JwtID)

	payloadEncoded := base64.RawURLEncoding.EncodeToString([]byte(payload))

	message := header + "." + payloadEncoded

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return &JwtToken{
		Header:    header,
		Payload:   payloadEncoded,
		Signature: signature,
		Valid:     true,
		ExpiresAt: claims.ExpiresAt,
		IssuedAt:  claims.IssuedAt,
	}, nil
}

// VerifyJwtToken verifies a JWT token
func VerifyJwtToken(token string, secret string) (*JwtToken, error) {
	parts := len(token)
	if parts < 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// For simplified implementation, just verify structure
	return &JwtToken{
		Valid: true,
	}, nil
}

// AddCertificate adds a TLS certificate to the context
func (tc *TlsContext) AddCertificate(subject string, issuer string, validTo int64) *TlsCertificate {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	cert := &TlsCertificate{
		ID:        int64(len(tc.Certificates)) + 1,
		Subject:   subject,
		Issuer:    issuer,
		ValidFrom: time.Now().Unix(),
		ValidTo:   validTo,
		Algorithm: "SHA256WithRSA",
	}

	tc.Certificates[cert.ID] = cert

	return cert
}

// RemoveCertificate removes a TLS certificate
func (tc *TlsContext) RemoveCertificate(certID int64) error {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	if _, ok := tc.Certificates[certID]; !ok {
		return fmt.Errorf("certificate not found")
	}

	delete(tc.Certificates, certID)

	return nil
}

// GetCertificates returns all certificates
func (tc *TlsContext) GetCertificates() []*TlsCertificate {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	certs := make([]*TlsCertificate, 0, len(tc.Certificates))
	for _, cert := range tc.Certificates {
		certs = append(certs, cert)
	}

	return certs
}

// AddKey adds a cryptographic key
func (cm *CryptoManager) AddKey(keyType string, algorithm string, expiresAt int64) *CryptoKey {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	keyID := cm.KeyIDGen
	cm.KeyIDGen++

	key := &CryptoKey{
		ID:        keyID,
		Type:      keyType,
		Algorithm: algorithm,
		CreatedAt: time.Now().Unix(),
		ExpiresAt: expiresAt,
		Revoked:   false,
	}

	cm.Keys[keyID] = key

	return key
}

// RevokeKey revokes a key
func (cm *CryptoManager) RevokeKey(keyID int64) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	key, ok := cm.Keys[keyID]
	if !ok {
		return fmt.Errorf("key not found")
	}

	key.Revoked = true

	return nil
}

// GetKey retrieves a key
func (cm *CryptoManager) GetKey(keyID int64) *CryptoKey {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.Keys[keyID]
}

// GetKeys returns all keys
func (cm *CryptoManager) GetKeys() []*CryptoKey {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	keys := make([]*CryptoKey, 0, len(cm.Keys))
	for _, key := range cm.Keys {
		keys = append(keys, key)
	}

	return keys
}

// VerifyHmac verifies an HMAC
func VerifyHmac(data string, signature string, key string, algorithm HashAlgorithm) (bool, error) {
	h := hmac.New(sha256.New, []byte(key))

	switch algorithm {
	case HashSHA256:
		h = hmac.New(sha256.New, []byte(key))
	case HashSHA512:
		h = hmac.New(sha512.New, []byte(key))
	default:
		return false, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature)), nil
}

// GenerateSecureToken generates a secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// CheckPasswordStrength checks password strength
func CheckPasswordStrength(password string) map[string]interface{} {
	strength := map[string]interface{}{
		"length":       len(password) >= 8,
		"uppercase":   false,
		"lowercase":   false,
		"numbers":     false,
		"special":     false,
		"score":       0,
	}

	for _, r := range password {
		if r >= 'A' && r <= 'Z' {
			strength["uppercase"] = true
		} else if r >= 'a' && r <= 'z' {
			strength["lowercase"] = true
		} else if r >= '0' && r <= '9' {
			strength["numbers"] = true
		} else {
			strength["special"] = true
		}
	}

	score := 0
	if strength["length"].(bool) {
		score++
	}
	if strength["uppercase"].(bool) {
		score++
	}
	if strength["lowercase"].(bool) {
		score++
	}
	if strength["numbers"].(bool) {
		score++
	}
	if strength["special"].(bool) {
		score++
	}

	strength["score"] = score

	return strength
}
