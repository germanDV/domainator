package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/germandv/domainator/internal/cache"
)

type DBPinger interface {
	Ping(ctx context.Context) error
}

func GetHealthcheck(cacheClient cache.Client, db DBPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		deep := q.Get("deep")

		var revision string
		var dirty bool
		var lastCommit time.Time
		info, ok := debug.ReadBuildInfo()
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Error reading build info"))
			return
		}

		for _, data := range info.Settings {
			switch data.Key {
			case "vcs.revision":
				revision = data.Value
			case "vcs.modified":
				dirty = true
			case "vcs.time":
				lastCommit, _ = time.Parse(time.RFC3339, data.Value)
			}
		}

		resp := map[string]any{
			"revision":   revision,
			"dirty":      dirty,
			"lastCommit": lastCommit,
			"go":         info.GoVersion,
		}

		if deep == "true" {
			cacheStatus := "up"
			err := cacheClient.Ping()
			if err != nil {
				cacheStatus = "down"
			}
			resp["redis"] = cacheStatus

			dbStatus := "up"
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			err = db.Ping(ctx)
			if err != nil {
				dbStatus = "down"
			}
			resp["postgres"] = dbStatus
		}

		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		_ = enc.Encode(resp)
	}
}
