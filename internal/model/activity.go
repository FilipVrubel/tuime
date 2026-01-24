package model

import "time"

type Activity struct {
	ID 	  int
	Name      string
	CreatedAt time.Time
}