package main

import (
	"bufio"
	"embed"
	"net"

	"gossip_common"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var (
	version           = "0.1.0"                 // Version of the application
	debugLogging      = false                   // Flag to enable debug logging
	connectionLogging bool                      // Flag to enable connection logging
	secureMode        bool                      // Flag to enable secure mode
	host              string                    // Host for signaling server
	port              int                       // Port for signaling server
	peerID            string                    // ID of the peer to connect to
	username          string                    // Username for the client
	publicKeys        = make(map[string][]byte) // Array of public keys from connected clients
	callID            = ""                      // ID of the call
	acceptedCallers   = make(map[string]bool)   // Map of accepted callers
	inCall            = false                   // Flag to check if the client is in a call
	conn              net.Conn                  // Connection to the signaling server
	writer            *bufio.Writer             // Writer for the connection
	password          string                    // Password for the server
	serverName        string                    // Name of the server
	muted             = false                   // Flag to check if the client is muted
	deafened          = false                   // Flag to check if the client is deafened
)

func main() {
	// Create an instance of the app structure
	app := NewApp()

	gossip_common.GenerateID(36)
	gossip_common.GenerateKeys()

	if debugLogging {
		gossip_common.Dbg("Client ID: %s", gossip_common.GetClientID())
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "Gossip",
		Width:     1024,
		Height:    768,
		MinWidth:  1024,
		MinHeight: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose:    app.cleanup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
