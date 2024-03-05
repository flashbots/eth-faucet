package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) renderJSON(w http.ResponseWriter, code int, v any) error {
	w.Header().Set(
		"Content-Type", "application/json",
	)
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}
