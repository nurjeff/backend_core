package tools

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GinRoute struct {
	Method     string
	Endpoint   string
	Handler    gin.HandlerFunc
	Permission uint
}

const (
	PERM_LOGIN = 0
	PERM_ZERO  = 1
	PERM_ADMIN = 2
)

type ErrorMessage struct {
	Code      int    `json:"code"`
	Error     string `json:"error"`
	Localized string `json:"localized"`
}

type Model struct {
	//ModelBasic
	ID        ModelID        `json:"id" gorm:"primaryKey;autoIncrement" update:"false"`
	CreatedAt time.Time      `json:"created_at" update:"false"`
	UpdatedAt time.Time      `json:"updated_at" update:"false"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index" update:"false"`
}

type Paging struct {
	Page       uint64      `json:"page"`
	TotalCount int64       `json:"total_count"`
	PerPage    uint64      `json:"per_page"`
	Data       interface{} `json:"data"`
}
