package model

// UserStar representation of user's starred repositories
type UserStar struct {
	User      string `json:"user,omitempty"`
	StarredAt int64  `json:"starredAt,omitempty"`
}

// Repository representation
type Repository struct {
	Name      string     `json:"name,omitempty"`
	Labels    []string   `json:"labels,omitempty"`
	Stars     []UserStar `json:"stars,omitempty"`
	Languages []string   `json:"languages,omitempty"`
}
