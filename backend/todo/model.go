package todo

import "time"

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
