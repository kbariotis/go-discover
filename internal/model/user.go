package model

type StarredRepository struct {
	Repository string
	StarredAt  int64
}

// User representation
type User struct {
	Name      string              `json:"name"`
	Followees []string            `json:"followees"`
	Stars     []StarredRepository `json:"stars"`
}
