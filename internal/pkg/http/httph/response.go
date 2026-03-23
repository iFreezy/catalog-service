package httph

import (
	"encoding/json"
	"net/http"
)

func SendJSON(w http.ResponseWriter, status int, data interface{}) {
	buf, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(buf)
}

func SendEmpty(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

func SendError(w http.ResponseWriter, status int, err error) {
	buf, marshalErr := json.Marshal(map[string]string{"error": err.Error()})
	if marshalErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(buf)
}
