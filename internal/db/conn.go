package db

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/model"
	"github.com/RacoonMediaServer/rms-packages/pkg/configuration"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	conn *gorm.DB
}

func Connect(config configuration.Database) (*Database, error) {
	db, err := gorm.Open(postgres.Open(config.GetConnectionString())) //&gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		return nil, err
	}
	if err = db.AutoMigrate(&model.Artist{}, &model.Content{}, &model.Torrent{}, &model.Chat{}); err != nil {
		return nil, err
	}
	//db.Migrator().CreateConstraint(&model.Artist{}, "Content")
	//db.Migrator().CreateConstraint(&model.Content{}, "Torrent")
	return &Database{conn: db}, nil
}
