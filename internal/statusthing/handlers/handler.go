package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"text/template"

	"github.com/lusis/apithings/internal/statusthing/providers"
	"github.com/lusis/apithings/internal/statusthing/types"

	"golang.org/x/exp/slog"
)

// StatusThingHandler is a struct that provides an http access for statusthings
type StatusThingHandler struct {
	provider      providers.Provider
	itemPathRegex *regexp.Regexp
	allPath       string
	logger        *slog.Logger
	apikey        string
	enableDash    bool
}

type httpRepresentation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

const thingIDRegexPattern = "([^/]+)"
const applicationJSON = "application/json"
const contentTypeHeader = "content-type"

// NewStatusThingHandler returns a new statusthing handler
func NewStatusThingHandler(provider providers.Provider, basePath string, logger *slog.Logger, apikey string, enableDash bool) (*StatusThingHandler, error) {
	if logger == nil {
		logger = slog.Default()
	}
	if basePath == "" {
		return nil, fmt.Errorf("basepath cannot be missing")
	}
	pathstring := fmt.Sprintf("^%s%s$", basePath, thingIDRegexPattern)
	pr := regexp.MustCompile(pathstring)
	return &StatusThingHandler{provider: provider, itemPathRegex: pr, allPath: basePath, logger: logger, apikey: apikey, enableDash: enableDash}, nil
}

func (h *StatusThingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.enableDash {
		if r.URL.Path == path.Join(h.allPath, "ui") {
			h.ui(r.Context(), w)
			return
		}
	}
	if h.apikey != "" {
		if r.Header.Get("X-STATUSTHING-KEY") != h.apikey {
			http.Error(w, "apikey required", http.StatusForbidden)
			return
		}
	}

	if r.Header.Get(contentTypeHeader) != applicationJSON {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	w.Header().Set(contentTypeHeader, applicationJSON)
	// short circuit if all request
	if r.URL.Path == h.allPath && r.Method == http.MethodGet {
		h.getall(r.Context(), w)
		return
	}

	matched := h.itemPathRegex.FindStringSubmatch(r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		if matched == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		id := matched[1]
		h.get(r.Context(), id, w)
	case http.MethodPost:
		if matched == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		id := matched[1]
		h.post(r.Context(), id, r.Body, w)
	case http.MethodDelete:
		if matched == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		id := matched[1]
		h.delete(r.Context(), id, w)
	case http.MethodPut:
		h.put(r.Context(), r.Body, w)
	default:
		http.Error(w, "unsupported method", http.StatusMethodNotAllowed)
	}
}

func (h *StatusThingHandler) ui(ctx context.Context, w http.ResponseWriter) {
	all, err := h.provider.All(ctx)
	if err != nil {
		h.logger.Error("error getting all results", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	t, err := template.New("dash").Parse(tmpl)
	if err != nil {
		h.logger.Error("error getting all results", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	cards := []string{}
	for _, thing := range all {
		cards = append(cards, makeCard(thing))
	}
	if err := t.Execute(w, cards); err != nil {
		h.logger.Error("error getting all results", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

func (h *StatusThingHandler) getall(ctx context.Context, w http.ResponseWriter) {
	all, err := h.provider.All(ctx)
	if err != nil {
		h.logger.Error("error getting all results", "err", err)
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
		h.logger.Error("encoding error", "err", err)
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
		h.logger.Error("unexpected error", "err", err)
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&httpRepresentation{
		ID:          res.ID,
		Name:        res.Name,
		Description: res.Description,
		Status:      res.Status.String(),
	}); err != nil {
		h.logger.Error("encoding error", "err", err)
		http.Error(w, "encoding error", http.StatusInternalServerError)
		return
	}
}

// Put provides a mechanism for adding a statusthing
func (h *StatusThingHandler) put(ctx context.Context, body io.ReadCloser, w http.ResponseWriter) {
	var entry = httpRepresentation{}
	if err := json.NewDecoder(body).Decode(&entry); err != nil {
		h.logger.Error("decoding error", "err", err)
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
		h.logger.Error("internal error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&httpRepresentation{
		ID:          res.ID,
		Name:        res.Name,
		Description: res.Description,
		Status:      res.Status.String(),
	}); err != nil {
		h.logger.Error("internal error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

// Post provides a mechanism for updating a statusthing
func (h *StatusThingHandler) post(ctx context.Context, id string, body io.ReadCloser, w http.ResponseWriter) {
	var entry = httpRepresentation{}
	if err := json.NewDecoder(body).Decode(&entry); err != nil {
		h.logger.Error("decoding error", "err", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.provider.SetStatus(ctx, id, types.StatusFromString(entry.Status)); err != nil {
		h.logger.Error("error setting status", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

// Delete provides a mechanism for deleting a statusthing
func (h *StatusThingHandler) delete(ctx context.Context, id string, w http.ResponseWriter) {
	existing, err := h.provider.Get(ctx, id)
	if errors.Is(err, types.ErrNotFound) {
		http.Error(w, "no such record", http.StatusNotFound)
		return
	}
	if err != nil {
		h.logger.Error("error getting existing record", "err", err)
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}
	if err := h.provider.Remove(ctx, id); err != nil {
		h.logger.Error("error removing entry", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&httpRepresentation{
		ID:          existing.ID,
		Name:        existing.Name,
		Description: existing.Description,
		Status:      existing.Status.String(),
	}); err != nil {
		h.logger.Error("internal error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}
