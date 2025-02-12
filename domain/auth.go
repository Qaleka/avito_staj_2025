package domain

import "context"

type User struct {
	UUID     string `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:uuid" json:"id"`
	Username string `gorm:"type:varchar(50);unique;not null;column:username" json:"username"`
	Password string `gorm:"type:varchar(255);not null;column:password" json:"password"`
	Coins    int    `gorm:"type:int;default:0;column:coins" json:"coins"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthRepository interface {
	AuthUser(ctx context.Context, username string, password string) (*User, error)
}
