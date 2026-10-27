package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "Mercurio",
		Width:     1100,
		Height:    820,
		Frameless: true, // custom title bar lives in the Svelte header
		MinWidth:  760,
		MinHeight: 600,
		// En Linux, con 0 Wails fija el máximo al tamaño del monitor donde arrancó
		// y la ventana no crece al moverla a otro más grande.
		MaxWidth:  16384,
		MaxHeight: 16384,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 23, G: 23, B: 33, A: 255},
		DragAndDrop:      &options.DragAndDrop{EnableFileDrop: true},
		Linux: &linux.Options{
			Icon:        icon,
			ProgramName: "mercurio-conversor", // Wayland app_id / WM class; matches the .desktop file
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
