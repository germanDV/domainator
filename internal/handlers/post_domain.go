package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/germandv/domainator/internal/templates"
)

func RegisterDomain(w http.ResponseWriter, r *http.Request) {
	domain := strings.TrimSpace(r.FormValue("domain"))

	if domain == "" || !strings.HasSuffix(domain, ".xyz") {
		w.WriteHeader(400)
		msg := fmt.Sprintf("%q is not a valid domain", domain)
		err := templates.RegisterDomainError(msg).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
		return
	}

	err := templates.RegisterDomainSuccess(domain).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
