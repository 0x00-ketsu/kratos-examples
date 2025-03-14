package model

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// BaseModel is the base model for all tables.
type BaseModel struct {
	ID        string                `json:"id" gorm:"primaryKey;size:32"`
	CreatedAt time.Time             `json:"createdAt" gorm:"<-:create"`
	UpdatedAt time.Time             `json:"updatedAt"`
	IsDeleted soft_delete.DeletedAt `json:"-" gorm:"softDelete:flag;size:8"`
}

func (b *BaseModel) BeforeUpdate(tx *gorm.DB) (err error) {
	b.UpdatedAt = time.Now()
	return
}
