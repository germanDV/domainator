// Package httphelp provides helper functions for common tasks around HTTP handling.
package httphelp

import (
	"domainator/internal/logger"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-playground/form/v4"
	"github.com/google/uuid"
)

type contextKey string

// UserIDContextKey is the key used to store the user ID in context.
const UserIDContextKey = contextKey("userID")

// PlanIDContextKey is the key used to store the subscription plan ID in context.
const PlanIDContextKey = contextKey("planID")

// formDecoder is global because it has a cache, so it's more efficient to reuse it.
var formDecoder = form.NewDecoder()

// ServerError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	logger.Writer.Error(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// ClientError helper sends a specific status code and corresponding description to the user.
func ClientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// NotFound is simply a convenience wrapper around ClientError which sends a 404 to the user.
func NotFound(w http.ResponseWriter) {
	ClientError(w, http.StatusNotFound)
}

// GetUserIDFromCtx returns the user ID from the request context.
func GetUserIDFromCtx(w http.ResponseWriter, r *http.Request) uuid.UUID {
	userID, ok := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if !ok || userID == uuid.Nil || userID.String() == "" {
		return uuid.Nil
	}
	return userID
}

// GetPlanIDFromCtx returns the subscribed plan ID from the request context.
func GetPlanIDFromCtx(w http.ResponseWriter, r *http.Request) int {
	planID, ok := r.Context().Value(PlanIDContextKey).(int)
	if !ok {
		return 0
	}
	return planID
}

// DecodeForm is a helper that decodes the form data from the request into the destination struct.
func DecodeForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}
	err = formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		return err
	}
	return nil
}
