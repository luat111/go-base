package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// Config holds PEM encoded RSA key material used by the Service.
type CryptoConfig struct {
	PublicKeyPEM  string `mapstructure:"PUBLIC_KEY_PEM"`
	PrivateKeyPEM string `mapstructure:"PRIVATE_KEY_PEM"`
}

// Service provides AES and RSA helpers used by the payload middleware.
type Service struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

// NewService builds a Service from PEM encoded key material.
func NewService(cfg CryptoConfig) (*Service, error) {
	if cfg.PublicKeyPEM == "" && cfg.PrivateKeyPEM == "" {
		return nil, errors.New("crypto: at least one RSA key is required")
	}

	svc := &Service{}

	if cfg.PublicKeyPEM != "" {
		pubKey, err := parsePublicKey(cfg.PublicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("crypto: parse public key: %w", err)
		}
		svc.publicKey = pubKey
	}

	if cfg.PrivateKeyPEM != "" {
		privKey, err := parsePrivateKey(cfg.PrivateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("crypto: parse private key: %w", err)
		}
		svc.privateKey = privKey
	}

	return svc, nil
}

// EncryptAES encrypts data with AES-256-CBC using a password-derived key and IV.
func (s *Service) EncryptAES(data []byte, password string) ([]byte, error) {
	key := deriveKey(password, 32)
	iv := deriveKey(password, aes.BlockSize)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	padded := pkcs7Pad(data, block.BlockSize())
	ciphertext := make([]byte, len(padded))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	return ciphertext, nil
}

// DecryptAES decrypts AES-256-CBC encrypted data using the provided password.
func (s *Service) DecryptAES(data []byte, password string) ([]byte, error) {
	key := deriveKey(password, 32)
	iv := deriveKey(password, aes.BlockSize)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(data)%block.BlockSize() != 0 {
		return nil, errors.New("crypto: invalid ciphertext length")
	}

	plaintext := make([]byte, len(data))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, data)

	unpadded, err := pkcs7Unpad(plaintext, block.BlockSize())
	if err != nil {
		return nil, err
	}

	return unpadded, nil
}

// EncryptRSA encrypts data using the configured public key and RSA OAEP padding.
func (s *Service) EncryptRSA(data []byte) ([]byte, error) {
	if s.publicKey == nil {
		return nil, errors.New("crypto: public key not configured")
	}

	return rsa.EncryptOAEP(sha1.New(), rand.Reader, s.publicKey, data, nil)
}

// DecryptRSA decrypts data using the configured private key and RSA OAEP padding.
func (s *Service) DecryptRSA(data []byte) ([]byte, error) {
	if s.privateKey == nil {
		return nil, errors.New("crypto: private key not configured")
	}

	return rsa.DecryptOAEP(sha1.New(), rand.Reader, s.privateKey, data, nil)
}

func parsePublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("crypto: invalid public key format")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("crypto: not an RSA public key")
	}

	return rsaKey, nil
}

func parsePrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("crypto: invalid private key format")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return privKey, nil
	}

	parsed, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err2 != nil {
		return nil, err
	}

	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("crypto: not an RSA private key")
	}

	return rsaKey, nil
}

func deriveKey(password string, length int) []byte {
	key := make([]byte, length)

	if password == "" {
		return key
	}

	src := []byte(password)
	for i := 0; i < length; i++ {
		key[i] = src[i%len(src)]
	}

	return key
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytesRepeat(byte(padding), padding)
	return append(data, padText...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 || length%blockSize != 0 {
		return nil, errors.New("crypto: invalid padding")
	}

	padding := int(data[length-1])
	if padding == 0 || padding > blockSize {
		return nil, errors.New("crypto: invalid padding size")
	}

	for i := 0; i < padding; i++ {
		if data[length-1-i] != byte(padding) {
			return nil, errors.New("crypto: invalid padding bytes")
		}
	}

	return data[:length-padding], nil
}

func bytesRepeat(b byte, count int) []byte {
	buf := make([]byte, count)
	for i := 0; i < count; i++ {
		buf[i] = b
	}
	return buf
}
