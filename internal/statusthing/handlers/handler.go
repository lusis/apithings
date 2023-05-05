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
	"strings"
	"text/template"

	"github.com/lusis/apithings/internal/statusthing/providers"
	"github.com/lusis/apithings/internal/statusthing/types"

	"golang.org/x/exp/slog"
)

const (
	// DefaultBasePath is the default base path
	DefaultBasePath = "/statusthings"
)

// StatusThingHandler is a struct that provides an http access for statusthings
type StatusThingHandler struct {
	provider      providers.Provider
	itemPathRegex *regexp.Regexp
	// basePath is the path where the handler is mounted
	basePath   string
	allPath    string
	apiPath    string
	apikey     string
	enableDash bool
}

type httpRepresentation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

const thingIDRegexPattern = "/([^/]+)"
const apiPathFragment = "/api/"
const applicationJSON = "application/json"
const contentTypeHeader = "content-type"

// NewStatusThingHandler returns a new statusthing handler
func NewStatusThingHandler(provider providers.Provider, opts ...HandlerOption) (*StatusThingHandler, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider cannot be nil")
	}
	sth := &StatusThingHandler{
		basePath: DefaultBasePath,
		provider: provider,
	}

	for _, opt := range opts {
		if err := opt(sth); err != nil {
			return nil, err
		}
	}
	sth.apiPath = path.Join(sth.basePath, apiPathFragment)
	apiPathstring := fmt.Sprintf("^%s%s$", sth.apiPath, thingIDRegexPattern)
	pr := regexp.MustCompile(apiPathstring)
	sth.itemPathRegex = pr
	return sth, nil
}

func (h *StatusThingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// serve the ui off the basepath. api is off /api/
	if r.URL.Path == h.basePath {
		h.ui(r.Context(), w)
		return
	}

	// we are now into the api routing. if it's not an api request it's a 404
	if !strings.HasPrefix(r.URL.Path, h.apiPath) {
		// if it's not the api path, 404
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// if the apikey is set and not provided, 403
	if h.apikey != "" {
		if r.Header.Get("X-STATUSTHING-KEY") != h.apikey {
			http.Error(w, "permission denied", http.StatusForbidden)
			return
		}
	}

	// for post/put we require body be json
	if r.Header.Get(contentTypeHeader) != applicationJSON && (r.Method == http.MethodPost || r.Method == http.MethodPut) {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	w.Header().Set(contentTypeHeader, applicationJSON)
	// we can short-circuit here if the path is the apipath
	isAPIRoot := (r.URL.Path == h.apiPath || r.URL.Path == h.apiPath+"/")
	if isAPIRoot && r.Method == http.MethodGet {
		h.getall(r.Context(), w)
		return
	}

	// Put requests to the root add new entries
	if isAPIRoot && r.Method == http.MethodPut {
		h.put(r.Context(), r.Body, w)
		return
	}

	// now we're into matching against the id for operations on those items
	matched := h.itemPathRegex.FindStringSubmatch(r.URL.Path)

	switch r.Method {
	// get the item with the provided id
	case http.MethodGet:
		if matched == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		id := matched[1]
		h.get(r.Context(), id, w)
	// update the item with the provided id
	case http.MethodPost:
		if matched == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		id := matched[1]
		h.post(r.Context(), id, r.Body, w)
	// remove the item with the provided id
	case http.MethodDelete:
		if matched == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		id := matched[1]
		h.delete(r.Context(), id, w)
	default:
		http.Error(w, "", http.StatusNotFound)
	}
}

// ui renders the super basic bootstrap dashboard
func (h *StatusThingHandler) ui(ctx context.Context, w http.ResponseWriter) {
	all, err := h.provider.All(ctx)
	if err != nil {
		slog.ErrorCtx(ctx, "error getting all results", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	t, err := template.New("dash").Parse(tmpl)
	if err != nil {
		slog.ErrorCtx(ctx, "error getting all results", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	cards := []string{}
	for _, thing := range all {
		cards = append(cards, makeCard(thing))
	}
	if err := t.Execute(w, cards); err != nil {
		slog.ErrorCtx(ctx, "error getting all results", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
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

// put provides a mechanism for adding a statusthing
func (h *StatusThingHandler) put(ctx context.Context, body io.ReadCloser, w http.ResponseWriter) {
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
		slog.ErrorCtx(ctx, "internal error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

// post provides a mechanism for updating a statusthing
func (h *StatusThingHandler) post(ctx context.Context, id string, body io.ReadCloser, w http.ResponseWriter) {
	var entry = httpRepresentation{}
	if err := json.NewDecoder(body).Decode(&entry); err != nil {
		slog.ErrorCtx(ctx, "decoding error", "err", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.provider.SetStatus(ctx, id, types.StatusFromString(entry.Status)); err != nil {
		slog.ErrorCtx(ctx, "error setting status", "err", err)
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
		slog.ErrorCtx(ctx, "error getting existing record", "err", err)
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
		slog.ErrorCtx(ctx, "internal error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}
