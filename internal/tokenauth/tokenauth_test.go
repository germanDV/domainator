package tokenauth

import (
	"testing"
	"time"

	"github.com/germandv/domainator/internal/keys"
	"github.com/google/uuid"
)

func TestTokenAuth(t *testing.T) {
	priv, publ, err := keys.NewPair()
	if err != nil {
		t.Fatal(err)
	}

	tokenAuth, err := New(priv, publ)
	if err != nil {
		t.Fatal(err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	userID := id.String()

	token, err := tokenAuth.Generate(userID)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := tokenAuth.Validate(token)
	if err != nil {
		t.Fatal(err)
	}

	if claims["sub"] != userID {
		t.Errorf("expected %s, got %s", userID, claims["sub"])
	}

	if claims["iss"] != "domainator" {
		t.Errorf("expected %s, got %s", "domainator", claims["iss"])
	}

	if claims["aud"] != "domainator" {
		t.Errorf("expected %s, got %s", "domainator", claims["aud"])
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		t.Errorf("expected float64, got %T", claims["iat"])
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Errorf("expected int64, got %T", claims["exp"])
	}

	if time.Unix(int64(iat), 0).Add(TokenExpiration).Unix() != int64(exp) {
		t.Error("iat claim + TokenExpiration does not match exp claim")
	}
}
