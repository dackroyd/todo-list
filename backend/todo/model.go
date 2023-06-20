package todo

import "time"

// NotFoundError occurs when trying to retrieve a specific value, and no such value exists.
type NotFoundError string

func (n NotFoundError) Error() string {
	return string(n)
}

type List struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type Item struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Due         *time.Time `json:"due"`
	Completed   *time.Time `json:"completed"`
}
