package windows

import (
	"github.com/gopal-lohar/hackathon-2025/internal/api/db"
	"github.com/gopal-lohar/hackathon-2025/internal/shared/utils/logger"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Windows struct {
	DB     *gorm.DB
	logger *logrus.Logger
}

func NewWindows() *Windows {
	logger := logger.NewLogger()
	db, err := db.NewDB()
	if err != nil {
		logger.Fatalf("Error connecting to db: %v", err)
	}
	return &Windows{
		DB:     db,
		logger: logger,
	}
}
