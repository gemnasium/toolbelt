package models

import "time"

type Alert struct {
	ID       int       `json:"id"`
	Advisory Advisory  `json:"advisory"`
	OpenAt   time.Time `json:"open_at"`
	Status   string    `json:"status"`
}
