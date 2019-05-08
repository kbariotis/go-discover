package model

// User representation
type User struct {
	Name      string   `json:"name"`
	Followees []string `json:"followees"`
	Stars     []string `json:"stars"`
}
