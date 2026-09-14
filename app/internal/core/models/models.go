package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type URLModel struct {
	Short_code   string `json:"short_code"`
	Original_url string `json:"original_url"`
}

type OriginalURL struct {
	Original string `json:"url"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	HashPass string `json:"-"`
}

type NewUser struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *User) CompareHash(IncomeHash []byte) error {
	return bcrypt.CompareHashAndPassword([]byte(u.HashPass), IncomeHash)
}

type TokenPair struct {
	UserID     int           `json:"user_id"`
	Access     string        `json:"access"`
	Refresh    string        `json:"refresh"`
	AccessTTL  time.Duration `json:"access_ttl"`
	RefreshTTL time.Duration `json:"refresh_ttl"`
}

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
