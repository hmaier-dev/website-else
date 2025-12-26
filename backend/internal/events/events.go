package events

import (
	"net/http"
	db "codeberg.org/schnelles_geraet/website-else/backend/internal/database"
)

type API struct {
	Username string
	Password string
}

type Handler struct{
	Queries *db.Queries
	API API
}

func (h *Handler) Display(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	return
}

func (h *Handler) AddEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	return
}

func (h *Handler) ChangeEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	return
}

func (h *Handler) RemoveEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	return
}
