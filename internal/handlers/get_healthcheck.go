package handlers

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/germandv/domainator/internal/cache"
)

func GetHealthcheck(cacheClient cache.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		var revision string
		var dirty bool
		var lastCommit time.Time
		info, ok := debug.ReadBuildInfo()
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error reading build info"))
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

		cacheStatus := "up"
		err := cacheClient.Ping()
		if err != nil {
			cacheStatus = "down"
		}

		resp := map[string]any{
			"revision":   revision,
			"dirty":      dirty,
			"lastCommit": lastCommit,
			"go":         info.GoVersion,
			"cache":      cacheStatus,
		}

		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		enc.Encode(resp)
	}
}
