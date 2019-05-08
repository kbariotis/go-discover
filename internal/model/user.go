package model

// User representation
type User struct {
	Name      string   `json:"name"`
	Followers []string `json:"followers"`
	Stars     []string `json:"stars"`
}
