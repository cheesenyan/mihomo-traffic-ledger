package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	readAutostartEnabled  = platformReadAutostartEnabled
	writeAutostartEnabled = platformWriteAutostartEnabled
)

type autostartSettings struct {
	Enabled bool `json:"enabled"`
}

func (s *service) handleAutostartSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		enabled, err := readAutostartEnabled()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, autostartSettings{Enabled: enabled})
	case http.MethodPut:
		var payload autostartSettings
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid json body"))
			return
		}
		if err := writeAutostartEnabled(payload.Enabled); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		enabled, err := readAutostartEnabled()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, autostartSettings{Enabled: enabled})
	default:
		writeMethodNotAllowed(w)
	}
}
