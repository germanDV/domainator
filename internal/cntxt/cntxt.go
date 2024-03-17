package cntxt

import (
	"context"
	"net/http"
)

type contextKey string

const contextKeyUserID = contextKey("userID")

func SetUserID(r *http.Request, userID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), contextKeyUserID, userID))
}

func GetUserID(r *http.Request) string {
	return r.Context().Value(contextKeyUserID).(string)
}
