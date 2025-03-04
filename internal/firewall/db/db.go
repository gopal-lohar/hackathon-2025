package db

import (
	"github.com/glebarez/sqlite"
	"github.com/gopal-lohar/hackathon-2025/internal/shared/schema"

	"gorm.io/gorm"
)

func NewDB() (*gorm.DB, error) {
	dbName := "firewall.sqlite3"
	database, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := database.DB()
	sqlDB.Exec("PRAGMA foreign_keys = ON")

	err = database.AutoMigrate(&schema.Endpoint{}, &schema.Rule{}, &schema.EndpointRule{})

	if err != nil {
		return nil, err
	}
	return database, nil
}
