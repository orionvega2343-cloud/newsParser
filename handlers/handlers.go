package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"newsParser/storage"
)

type Handler struct {
	DB *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	res, err := storage.GetAll(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
