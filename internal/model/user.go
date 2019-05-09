package model

import "github.com/kbariotis/go-discover/internal/provider"

// User representation
type User struct {
	Name      string                       `json:"name"`
	Followees []string                     `json:"followees"`
	Stars     []provider.StarredRepository `json:"stars"`
}
