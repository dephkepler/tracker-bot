package models

import "time"

type User struct {
	TgUserID    int64
	UserName    *string
	PhoneNumber *string
	Email       *string
	Language    *string
	TimeZone    *string
	CreatedAt   time.Time
}
