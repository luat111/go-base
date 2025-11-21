package encrypt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	defaultEncryptHeader = "X-Request-Encrypt-Key"

	ctxPasswordKey = "_crypto_password"
	ctxEnabledKey  = "_crypto_enabled"
)

// MiddlewareConfig controls the payload encryption middleware behaviour.
type MiddlewareConfig struct {
	HeaderName      string
	RequireHeader   bool
	DisableSanitize bool
	Skip            func(*gin.Context) bool
}

// PayloadMiddleware decrypts inbound requests (if needed) and re-encrypts responses.
func (s *Service) PayloadMiddleware(cfg MiddlewareConfig) gin.HandlerFunc {
	header := cfg.HeaderName
	if header == "" {
		header = defaultEncryptHeader
	}

	sanitize := !cfg.DisableSanitize

	return func(c *gin.Context) {
		if cfg.Skip != nil && cfg.Skip(c) {
			c.Next()
			return
		}

		encryptedPassword := strings.TrimSpace(c.GetHeader(header))
		if encryptedPassword == "" {
			if cfg.RequireHeader {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "encryption key required"})
				return
			}

			c.Next()
			return
		}

		password, err := s.decryptPassword(encryptedPassword)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid encryption key"})
			return
		}

		if err := s.decryptRequestBody(c, password, sanitize); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		writer := newPayloadWriter(c.Writer)
		c.Writer = writer

		c.Set(ctxPasswordKey, password)
		c.Set(ctxEnabledKey, true)

		c.Next()

		if !writer.capture {
			return
		}

		if err := s.encryptResponse(c, writer, password); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "encrypt response failed"})
			return
		}
	}
}

func (s *Service) decryptPassword(headerValue string) (string, error) {
	if s.privateKey == nil {
		return "", errors.New("crypto: private key not configured")
	}

	buf, err := base64.StdEncoding.DecodeString(headerValue)
	if err != nil {
		return "", err
	}

	plain, err := s.DecryptRSA(buf)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}

func (s *Service) decryptRequestBody(c *gin.Context, password string, sanitize bool) error {
	if c.Request.Body == nil {
		return nil
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}

	if len(bytes.TrimSpace(bodyBytes)) == 0 {
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return nil
	}

	cipherBody, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(bodyBytes)))
	if err != nil {
		return err
	}

	plainBody, err := s.DecryptAES(cipherBody, password)
	if err != nil {
		return err
	}

	if sanitize {
		plainBody, err = sanitizeJSON(plainBody)
		if err != nil {
			return err
		}
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(plainBody))
	c.Request.ContentLength = int64(len(plainBody))
	c.Request.Header.Set("Content-Length", strconv.Itoa(len(plainBody)))

	if c.ContentType() == "" {
		c.Request.Header.Set("Content-Type", "application/json")
	}

	return nil
}

func (s *Service) encryptResponse(c *gin.Context, writer *payloadWriter, password string) error {
	payload := writer.body.Bytes()

	cipherBody, err := s.EncryptAES(payload, password)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(cipherBody)
	rawWriter := writer.ResponseWriter

	rawWriter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rawWriter.Header().Set("Content-Length", strconv.Itoa(len(encoded)))

	status := writer.Status()
	if status == 0 {
		status = http.StatusOK
	}

	rawWriter.WriteHeader(status)
	_, err = rawWriter.Write([]byte(encoded))

	return err
}

type payloadWriter struct {
	gin.ResponseWriter
	body    bytes.Buffer
	capture bool
}

func newPayloadWriter(w gin.ResponseWriter) *payloadWriter {
	return &payloadWriter{
		ResponseWriter: w,
		capture:        true,
	}
}

func (w *payloadWriter) Write(data []byte) (int, error) {
	if w.capture {
		return w.body.Write(data)
	}

	return w.ResponseWriter.Write(data)
}

func (w *payloadWriter) WriteString(s string) (int, error) {
	if w.capture {
		return w.body.WriteString(s)
	}

	return w.ResponseWriter.WriteString(s)
}

func sanitizeJSON(body []byte) ([]byte, error) {
	var payload any

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	sanitized, err := sanitizeValue(payload)
	if err != nil {
		return nil, err
	}

	return json.Marshal(sanitized)
}

func sanitizeValue(value any) (any, error) {
	switch v := value.(type) {
	case string:
		if strings.ContainsRune(v, '\u0000') {
			return nil, errors.New("unrecognized_input_format")
		}
		return strings.ReplaceAll(v, "\u0000", ""), nil
	case []any:
		for i := range v {
			next, err := sanitizeValue(v[i])
			if err != nil {
				return nil, err
			}
			v[i] = next
		}
		return v, nil
	case map[string]any:
		for key := range v {
			next, err := sanitizeValue(v[key])
			if err != nil {
				return nil, err
			}
			v[key] = next
		}
		return v, nil
	default:
		return v, nil
	}
}
