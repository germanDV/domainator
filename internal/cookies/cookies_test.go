package cookies

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

const secret = "4yxfIPahS5s15puGIDIDFqVSm09mKkyH"

func TestWrite(t *testing.T) {
	// Create a new test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		cookie := http.Cookie{Name: "test", Value: "value"}
		err := Write(w, cookie)
		if err != nil {
			t.Errorf("Write returned an error: %v", err)
		}
	}))
	defer ts.Close()

	// Make a request to the test server
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	// Check if the cookie was set
	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Errorf("Expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "test" || cookies[0].Value != "value" {
		t.Errorf("Unexpected cookie value: %v", cookies[0])
	}
}

func TestRead(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	cookie := &http.Cookie{Name: "testCookie", Value: "testValue"}
	req.AddCookie(cookie)

	value, err := Read(req, "testCookie")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if value != "testValue" {
		t.Errorf("Expected value to be 'testValue', but got %s", value)
	}
}

func TestReadMissingCookie(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	_, err := Read(req, "missingCookie")
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}
}

func TestWriteEncoded(t *testing.T) {
	// Create a new test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		cookie := http.Cookie{Name: "test", Value: "value"}
		err := WriteEncoded(w, cookie)
		if err != nil {
			t.Errorf("Write returned an error: %v", err)
		}
	}))
	defer ts.Close()

	// Make a request to the test server
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	// Check if the cookie was set
	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Errorf("Expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "test" {
		t.Errorf("Unexpected cookie name: %v", cookies[0])
	}
	if cookies[0].Value != base64.URLEncoding.EncodeToString([]byte("value")) {
		t.Errorf("Unexpected cookie value: %v", cookies[0])
	}
}

func TestReadEncoded(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	cookie := &http.Cookie{Name: "testCookie", Value: "testValue"}
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))
	req.AddCookie(cookie)

	value, err := ReadEncoded(req, "testCookie")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if value != "testValue" {
		t.Errorf("Expected value to be 'testValue', but got %s", value)
	}
}

func TestWriteSigned(t *testing.T) {
	// Create a new test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		cookie := http.Cookie{Name: "test", Value: "value"}
		err := WriteSigned(w, cookie, []byte(secret))
		if err != nil {
			t.Errorf("Write returned an error: %v", err)
		}
	}))
	defer ts.Close()

	// Make a request to the test server
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}

	// Check if the cookie was set
	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Errorf("Expected 1 cookie, got %d", len(cookies))
	}

	if cookies[0].Name != "test" {
		t.Errorf("Unexpected cookie name: %v", cookies[0])
	}

	if len(cookies[0].Value) < sha256.Size {
		t.Errorf("Unexpected cookie value: %v", cookies[0])
	}

	decoded, err := base64.URLEncoding.DecodeString(cookies[0].Value)
	if err != nil {
		t.Errorf("Error decoding value: %v", err)
	}

	value := decoded[sha256.Size:]
	if string(value) != "value" {
		t.Errorf("Unexpected cookie value: %s, expected 'value'", string(value))
	}
}

func TestReadSigned(t *testing.T) {
	name := "cookie"
	value := "value"

	req, _ := http.NewRequest("GET", "/", nil)
	signedCookie := signCookie(name, value, []byte(secret))
	req.AddCookie(signedCookie)

	result, err := ReadSigned(req, name, []byte(secret))
	if err != nil {
		t.Errorf("Error reading signed cookie: %v", err)
	}
	if result != value {
		t.Errorf("Expected value %s, got %s", value, result)
	}
}

func TestReadSignedInvalidValue(t *testing.T) {
	name := "test"
	value := "original"

	req, _ := http.NewRequest("GET", "/", nil)
	signedCookie := signCookie(name, value, []byte(secret))

	// Tamper the cookie
	signedCookie.Value += "tampered"

	req.AddCookie(signedCookie)

	_, err := ReadSigned(req, name, []byte(secret))
	if err == nil {
		t.Errorf("Expected error for tampered value, got nil")
	}
}

func signCookie(name string, value string, secret []byte) *http.Cookie {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	signature := mac.Sum(nil)
	return &http.Cookie{
		Name:  name,
		Value: base64.URLEncoding.EncodeToString([]byte(string(signature) + value)),
	}
}
