package model

// StarredRepository representation of user's starred repositories
type StarredRepository struct {
	// gorm.Model
	ID         uint   `gorm:"primary_key"`
	Repository string `json:"repository,omitempty"`
	StarredAt  int64  `json:"starredAt,omitempty"`
}

// User representation
type User struct {
	// gorm.Model
	ID        uint                `gorm:"primary_key"`
	Name      string              `json:"name,omitempty"`
	Email     string              `json:"-" gorm:"column:email"`
	Followees []string            `json:"followees,omitempty" gorm:"-"`
	Stars     []StarredRepository `json:"stars,omitempty" gorm:"-"`
}
