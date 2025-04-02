package main

import (
	"context"
	"fmt"
	"infinity-subtitle/database"
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

	// Initialize database
	err := database.GetDB().Init()
	if err != nil {
		println("Error initializing database:", err.Error())
	}

	// Check if tables exist
	exists, err := database.CheckTablesExists()
	if err != nil {
		println("Error checking tables:", err.Error())
	}

	if !exists {
		err = database.CreateTables()
		if err != nil {
			println("Error creating tables:", err.Error())
		}
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
