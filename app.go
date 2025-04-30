package main

import (
	"context"
	"infinity-subtitle/backend/database"
	"infinity-subtitle/backend/logger"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logger, err := logger.GetLogger()
	if err != nil {
		logger.Error("Error initializing logger:", err.Error())
	}

	// Initialize database
	err = database.GetDB().Init();
	if err != nil {
		logger.Error("Error initializing database:", err.Error())
	}

	logger.Info("App started")

	// Check if tables exist and create them if they don't
	err = database.CheckTablesExists()
	if err != nil {
		logger.Error("Error checking tables:", err.Error())
	}
}