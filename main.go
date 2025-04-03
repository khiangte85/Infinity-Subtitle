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

	// Create application with options
	err := wails.Run(&options.App{
		Title:      "Infinity Subtitle",
		Fullscreen: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 173, G: 216, B: 230, A: 1},
		OnStartup:        app.startup,
		Bind: []any{
			app,
			language,
			movie,
		},
		AlwaysOnTop: false,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
