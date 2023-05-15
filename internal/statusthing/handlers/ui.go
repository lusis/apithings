package handlers

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/lusis/apithings/internal/statusthing/types"

	"golang.org/x/exp/slog"
)

const (
	bgSuccessCard = "bg-success"
	bgDangerCard  = "bg-danger"
	bgWarningCard = "bg-warning"
)

type card struct {
	Style string
	Title string
	ID    string
	Desc  string
}

func makeCard(thing *types.StatusThing) card {
	bgstring := "bg-primary"
	switch thing.Status {
	case types.StatusGreen:
		bgstring = bgSuccessCard
	case types.StatusYellow:
		bgstring = bgWarningCard
	case types.StatusRed:
		bgstring = bgDangerCard
	}
	name := template.HTMLEscapeString(thing.Name)
	desc := template.HTMLEscapeString(thing.Description)
	return card{
		Style: bgstring,
		Title: name,
		ID:    thing.ID,
		Desc:  desc,
	}
}

func (h *StatusThingHandler) addUIRoutes(r chi.Router) {
	r.Get("/cards", func(w http.ResponseWriter, r *http.Request) {
		all, err := h.provider.All(r.Context())
		if err != nil {
			slog.Error("error getting all results", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		cards := []card{}
		for _, thing := range all {
			cards = append(cards, makeCard(thing))
		}
		tmpl := h.templates["card.htmx"]

		if err := tmpl.Execute(w, cards); err != nil {
			slog.Error("error executing template", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := h.templates["index.htmx"]
		if err := tmpl.Execute(w, nil); err != nil {
			slog.Error("error executing template", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	})
}
