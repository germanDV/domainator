package cookies

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
)

var (
	ErrValueTooLong = errors.New("cookie value too long")
	ErrInvalidValue = errors.New("invalid cookie value")
)

// Write writes the cookie to the response writer.
func Write(w http.ResponseWriter, cookie http.Cookie) error {
	if len(cookie.String()) > 4096 {
		return ErrValueTooLong
	}
	http.SetCookie(w, &cookie)
	return nil
}

// Read reads the cookie from the request.
func Read(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

// WriteEncoded encodes the cookie value in base64 and writes it to the response.
func WriteEncoded(w http.ResponseWriter, cookie http.Cookie) error {
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))
	return Write(w, cookie)
}

// Read reads a base64-enconded cookie from the request and returns the decoded value.
func ReadEncoded(r *http.Request, name string) (string, error) {
	value, err := Read(r, name)
	if err != nil {
		return "", err
	}

	decoded, err := base64.URLEncoding.DecodeString(value)
	if err != nil {
		return "", ErrInvalidValue
	}

	return string(decoded), nil
}

// WriteSigned calculates a HMAC signature of the cookie name and value
// and prepends it to the cookie value.
func WriteSigned(w http.ResponseWriter, cookie http.Cookie, secret []byte) error {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(cookie.Name))
	mac.Write([]byte(cookie.Value))
	signature := mac.Sum(nil)
	cookie.Value = string(signature) + cookie.Value
	return WriteEncoded(w, cookie)
}

// ReadSigned reads the cookie from the request and verifies the HMAC signature.
func ReadSigned(r *http.Request, name string, secret []byte) (string, error) {
	signatureAndValue, err := ReadEncoded(r, name)
	if err != nil {
		return "", err
	}

	if len(signatureAndValue) < sha256.Size {
		return "", ErrInvalidValue
	}

	signature := signatureAndValue[:sha256.Size]
	value := signatureAndValue[sha256.Size:]

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)
	if !hmac.Equal([]byte(signature), expectedSignature) {
		return "", ErrInvalidValue
	}

	return value, nil
}
