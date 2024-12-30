package auth

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`
}

func (u *User) validate() []string {
	issues := []string{}
	if u.Username == "" {
		issues = append(issues, "`username` is required")
	}
	if u.Password == "" {
		issues = append(issues, "`password` is required")
	}

	return issues
}
