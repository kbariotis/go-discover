package model

// Repository representation
type Repository struct {
	Name      string   `json:"name"`
	Labels    []string `json:"labels"`
	Languages []string `json:"languages"`
}
