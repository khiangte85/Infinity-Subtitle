package main

import (
	"embed"
	"infinity-subtitle/backend"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()
	language := backend.NewLanguage()
	movie := backend.NewMovie()
	subtitle := backend.NewSubtitle()
	setting := backend.NewSetting()
	movieQueue := backend.NewMovieQueue()

	// Create application with options
	err := wails.Run(&options.App{
		Title:      "Infinity Subtitle",
		Fullscreen: false,
		Width:      1366,
		Height:     768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 173, G: 216, B: 230, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []any{
			app,
			language,
			movie,
			subtitle,
			setting,
			movieQueue,
		},
		AlwaysOnTop: false,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
