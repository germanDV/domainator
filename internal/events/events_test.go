package events

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var controller *Controller

func TestMain(m *testing.M) {
	setup()
	os.Exit(m.Run())
}

func setup() {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validate := validator.New()
	repo := &InMemoryRepo{db: make(map[uuid.UUID]Event)}
	controller = NewController(repo, validate, logger)
}

func mockRequest(eventName string) (*http.Request, error) {
	r, err := http.NewRequest(http.MethodPost, "/events", nil)
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("name", eventName)
	r.PostForm = form
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return r, nil
}

func TestSave(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		w := httptest.NewRecorder()

		eventName := "test_event"
		r, err := mockRequest(eventName)
		if err != nil {
			t.Fatal(err)
		}

		controller.Save(w, r)
		resp := w.Result()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("want status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != `<div class="center">Thanks for letting us know!</div>` {
			t.Errorf("response body does not contain the expected HTML: %s", string(body))
		}

		evs, err := controller.repo.GetAll(context.Background(), eventName)
		if err != nil {
			t.Fatal(err)
		}
		if len(evs) != 1 {
			t.Errorf("want 1 event, got %d", len(evs))
		}
		if evs[0].Name != eventName {
			t.Errorf("want %s, got %s", eventName, evs[0].Name)
		}
		if evs[0].UserID != uuid.Nil {
			t.Errorf("want userID to be nil UUID, got %s", evs[0].UserID)
		}
	})

	t.Run("invalid_payload", func(t *testing.T) {
		w := httptest.NewRecorder()

		tooLong := strings.Repeat("a", 65)
		r, err := mockRequest(tooLong)
		if err != nil {
			t.Fatal(err)
		}

		controller.Save(w, r)
		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("want status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("missing_required_field", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, "/events", nil)
		if err != nil {
			t.Fatal(err)
		}

		controller.Save(w, r)
		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("want status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})
}

// In-memory repository of events
type InMemoryRepo struct {
	db map[uuid.UUID]Event
}

func (r *InMemoryRepo) Save(_ context.Context, userID uuid.UUID, ev *CreateEventReq) (uuid.UUID, error) {
	id := uuid.New()
	r.db[id] = Event{
		ID:      id,
		UserID:  userID,
		Name:    ev.Name,
		Payload: ev.Payload,
	}
	return id, nil
}

func (r *InMemoryRepo) GetAll(_ context.Context, eventName string) ([]Event, error) {
	events := []Event{}
	for _, ev := range r.db {
		if ev.Name == eventName {
			events = append(events, ev)
		}
	}
	return events, nil
}
