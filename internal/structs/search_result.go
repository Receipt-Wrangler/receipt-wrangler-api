package structs

import "time"

type SearchResult struct {
	ID      uint      `json:"id"`
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	GroupID uint      `json:"groupId"`
	Date    time.Time `json:"date"`
}
