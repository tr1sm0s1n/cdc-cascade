package models

import "time"

type Sinner struct {
	Code      int       `json:"code" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Class     string    `json:"class"`
	Libram    string    `json:"libram"`
	Tendency  string    `json:"tendency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
