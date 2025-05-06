package main

import (
	"context"
	"infinity-subtitle/backend"
	"infinity-subtitle/backend/database"
	"infinity-subtitle/backend/logger"
	"sync"
	"time"
)

// App struct
type App struct {
	ctx        context.Context
	wg         sync.WaitGroup
	cancelFunc context.CancelFunc
	logger     *logger.Logger
}

// NewApp creates a new App application struct
func NewApp() *App {
	logger, err := logger.GetLogger()
	if err != nil {
		// If logger creation failed, we can only print to stderr since logger isn't available
		println("Failed to initialize logger:", err.Error())
	}
	return &App{logger: logger}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx, a.cancelFunc = context.WithCancel(ctx)

	// Initialize database
	err := database.GetDB().Init()

	if err != nil {
		a.logger.Error("Error initializing database:", err.Error())
	}

	a.logger.Info("App started")

	// Check if tables exist and create them if they don't
	err = database.CheckTablesExists()
	if err != nil {
		a.logger.Error("Error checking tables:", err.Error())
	}

	// Run CreateMovieFromQueue worker
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				a.logger.Info("Context cancelled, exiting createMovieFromQueue goroutine")
				return
			default:
				backend.CreateMovieFromQueue(a.ctx)
				time.Sleep(2 * time.Second)
			}
		}
	}()

	// Run CreateSubtitleFromQueue worker
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				a.logger.Info("Context cancelled, exiting createSubtitleFromQueue goroutine")
				return
			default:
				backend.CreateSubtitleFromQueue(a.ctx)
				time.Sleep(2 * time.Second)
			}
		}
	}()

	// Run TranslateSubtitleFromQueue workers
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				a.logger.Info("Context cancelled, exiting translateSubtitleFromQueue goroutine")
				return
			default:
				backend.TranslateSubtitleFromQueue(a.ctx)
				time.Sleep(2 * time.Second)
			}
		}
	}()

}

func (a *App) shutdown(ctx context.Context) {
	a.logger.Info("Shutting down app")

	a.cancelFunc()

	done := make(chan struct{})

	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("All goroutines finished")
	case <-ctx.Done():
		a.logger.Info("Shutdown timed out, forcing shutdown")
	}

}
