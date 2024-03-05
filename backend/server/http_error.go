package server

import "net/http"

func (s *Server) httpError(w http.ResponseWriter, httpError int) {
	http.Error(
		w,
		http.StatusText(httpError),
		httpError,
	)
}
