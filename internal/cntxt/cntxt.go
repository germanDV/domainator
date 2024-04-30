package cntxt

import (
	"context"
	"net/http"
)

type contextKey string

const (
	ContextKeyUserID = contextKey("userID")
	ContextKeyAvatar = contextKey("avatar")
)

func SetUserID(r *http.Request, userID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ContextKeyUserID, userID))
}

func GetUserID(r *http.Request) string {
	id, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok {
		return ""
	}
	return id
}

func SetAvatar(r *http.Request, avatar string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ContextKeyAvatar, avatar))
}

func GetAvatar(r *http.Request) string {
	avatar, ok := r.Context().Value(ContextKeyAvatar).(string)
	if !ok {
		return ""
	}
	return avatar
}
