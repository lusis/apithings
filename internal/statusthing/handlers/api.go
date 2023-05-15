package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	chi "github.com/go-chi/chi/v5"

	"github.com/lusis/apithings/internal/statusthing/providers"
	"github.com/lusis/apithings/internal/statusthing/types"

	"golang.org/x/exp/slog"
)

func (h *StatusThingHandler) addAPIRoutes(r chi.Router) {
	// add in some middleware for checking auth and content-type
	r.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// check for api key if required
			if h.apikey != "" {
				if r.Header.Get("X-STATUSTHING-KEY") != h.apikey {
					http.Error(w, "permission denied", http.StatusForbidden)
					return
				}
			}
			// check for content-type
			if r.Header.Get(contentTypeHeader) != applicationJSON {
				http.Error(w, "invalid content type", http.StatusBadRequest)
				return
			}
			handler.ServeHTTP(w, r)
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		h.getall(r.Context(), w)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(contentTypeHeader) != applicationJSON {
			http.Error(w, "invalid content type", http.StatusBadRequest)
			return
		}
		h.post(r.Context(), r.Body, w)
	})

	r.Put("/{thingID}", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(contentTypeHeader) != applicationJSON {
			http.Error(w, "invalid content type", http.StatusBadRequest)
			return
		}
		thingID := chi.URLParam(r, "thingID")
		h.put(r.Context(), thingID, r.Body, w)
	})

	r.Get("/{thingID}", func(w http.ResponseWriter, r *http.Request) {
		thingID := chi.URLParam(r, "thingID")
		h.get(r.Context(), thingID, w)
	})

	r.Delete("/{thingID}", func(w http.ResponseWriter, r *http.Request) {
		thingID := chi.URLParam(r, "thingID")
		h.delete(r.Context(), thingID, w)
	})
}

// getall gets all known things
func (h *StatusThingHandler) getall(ctx context.Context, w http.ResponseWriter) {
	all, err := h.provider.All(ctx)
	if err != nil {
		slog.ErrorCtx(ctx, "error getting all results", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	res := []*httpRepresentation{}
	for _, i := range all {
		res = append(res, &httpRepresentation{
			ID:          i.ID,
			Description: i.Description,
			Name:        i.Name,
			Status:      i.Status.String(),
		})
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		slog.ErrorCtx(ctx, "encoding error", "err", err)
		http.Error(w, "encoding error", http.StatusInternalServerError)
		return
	}
}

// get returns a statusthing by id
func (h *StatusThingHandler) get(ctx context.Context, id string, w http.ResponseWriter) {
	res, err := h.provider.Get(ctx, id)
	if err == types.ErrNotFound {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.ErrorCtx(ctx, "unexpected error", "err", err)
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&httpRepresentation{
		ID:          res.ID,
		Name:        res.Name,
		Description: res.Description,
		Status:      res.Status.String(),
	}); err != nil {
		slog.ErrorCtx(ctx, "encoding error", "err", err)
		http.Error(w, "encoding error", http.StatusInternalServerError)
		return
	}
}

// post provides a mechanism for adding a statusthing
func (h *StatusThingHandler) post(ctx context.Context, body io.ReadCloser, w http.ResponseWriter) {
	var entry = httpRepresentation{}
	if err := json.NewDecoder(body).Decode(&entry); err != nil {
		slog.ErrorCtx(ctx, "decoding error", "err", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	res, err := h.provider.Add(ctx, providers.Params{Name: entry.Name, Description: entry.Description, Status: types.StatusFromString(entry.Status)})
	if errors.Is(err, types.ErrRequiredValueMissing) {
		http.Error(w, fmt.Sprintf("validation failed: %s", err.Error()), http.StatusBadRequest)
		return
	}
	if errors.Is(err, types.ErrAlreadyExists) {
		http.Error(w, "service already exists with that name", http.StatusConflict)
		return
	}
	if err != nil {
		slog.ErrorCtx(ctx, "internal error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&httpRepresentation{
		ID:          res.ID,
		Name:        res.Name,
		Description: res.Description,
		Status:      res.Status.String(),
	}); err != nil {
		slog.Error("internal error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

// put provides a mechanism for updating a statusthing
func (h *StatusThingHandler) put(ctx context.Context, id string, body io.ReadCloser, w http.ResponseWriter) {
	var entry = httpRepresentation{}
	if err := json.NewDecoder(body).Decode(&entry); err != nil {
		slog.Error("decoding error", "err", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.provider.SetStatus(ctx, id, types.StatusFromString(entry.Status)); err != nil {
		slog.Error("error setting status", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

// delete provides a mechanism for deleting a statusthing
func (h *StatusThingHandler) delete(ctx context.Context, id string, w http.ResponseWriter) {
	existing, err := h.provider.Get(ctx, id)
	if errors.Is(err, types.ErrNotFound) {
		http.Error(w, "no such record", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("error getting existing record", "err", err)
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}
	if err := h.provider.Remove(ctx, id); err != nil {
		slog.ErrorCtx(ctx, "error removing entry", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&httpRepresentation{
		ID:          existing.ID,
		Name:        existing.Name,
		Description: existing.Description,
		Status:      existing.Status.String(),
	}); err != nil {
		slog.Error("internal error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}
