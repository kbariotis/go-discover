package model

// StarredRepository representation of user's starred repositories
type StarredRepository struct {
	Repository string `json:"repository,omitempty"`
	StarredAt  int64  `json:"starredAt,omitempty"`
}

// User representation
type User struct {
	Name      string              `json:"name,omitempty"`
	Followees []string            `json:"followees,omitempty"`
	Stars     []StarredRepository `json:"stars,omitempty"`
}
