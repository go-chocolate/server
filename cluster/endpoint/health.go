package endpoint

type Health struct {
	Protocol Protocol `json:"protocol"`
	Path     string   `json:"path"`
}
