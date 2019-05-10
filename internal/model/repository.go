package model

// Repository representation
type Repository struct {
	Name      string   `json:"name,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	Languages []string `json:"languages,omitempty"`
}
