package handlers

import (
	"log/slog"
	"net/http"

	"github.com/germandv/domainator/internal/cntxt"
	"github.com/germandv/domainator/internal/common"
	"github.com/germandv/domainator/internal/notifier"
	"github.com/germandv/domainator/internal/users"
)

func SendTestMessage(logger *slog.Logger, userService users.Service, n notifier.Notifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDstr := cntxt.GetUserID(r)
		userID, err := common.ParseID(userIDstr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		u, err := userService.GetByID(r.Context(), users.GetByIDReq{UserID: userID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		url := u.WebhookURL.String()
		if url == "" {
			http.Error(w, "Please make sure you have saved your webhook URL", http.StatusBadRequest)
			return
		}

		notification := notifier.Notification{
			ID:     "",
			UserID: u.ID.String(),
			Domain: "This is a Test Message",
			Status: "OK",
			Hours:  0,
		}

		err = n.Notify(url, notification)
		if err != nil {
			logger.Error("Failed to send test message", "error", err, "user", userID, "webhook", url)
			http.Error(w, "Error sending test message", http.StatusInternalServerError)
			return
		}

		logger.Info("Test message sent", "user", userID, "webhook", url)
		c := MessageSent()
		SendTempl(w, r, c)
	}
}
