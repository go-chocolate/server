package handler

import "net/http"

const (
	HealthPath = "/__health__"
)

func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("success"))
}
