package persistence

import "time"

type AuthDB struct {
	ID        string
	Email     string
	Password  string
	CreatedAt time.Time
}
