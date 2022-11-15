package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID `gorm:"primarykey;type:uuid;default:public.uuid_generate_v4()"`
	Name      string
	Surname   string
	Email     string `gorm:"index"`
	Password  string
	Role      string `gorm:"index;default:user"`
	CreatedAt time.Time
	IsDeleted int32 `gorm:"default:0;index"`
}
