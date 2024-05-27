package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gossip_common"
)

var (
	debugLogging      bool
	connectionLogging bool
	host              string
	port              int
	version           = "0.1.0"
	password          string
	channels          []string
	serverName        string
)

var (
	connections     = make(map[string]net.Conn)
	publicKeys      = make(map[string][]byte)
	connectionsLock sync.RWMutex
)

func init() {
	flag.BoolVar(&debugLogging, "d", false, "Enable debug logging")
	flag.BoolVar(&connectionLogging, "l", false, "Enable connection logging")
	flag.StringVar(&host, "h", "127.0.0.1", "IP to listen on")
	flag.IntVar(&port, "p", 1720, "Port to listen on")
	flag.StringVar(&password, "k", "anonymous", "Password for the server")
	flag.Parse()
}

func main() {
	gossip_common.Log("Starting gossip " + version + " server...")

	if debugLogging {
		gossip_common.Log("Debug logging enabled!")
	}

	if connectionLogging {
		gossip_common.Log("Connection logging enabled!")
	}

	if password != "" {
		password = gossip_common.HashPassword(password)
	}

	compileChannels()
	fetchName()

	gossip_common.GenerateKeys()

	addr := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		gossip_common.Err("Failed to listen on %s: %v", addr, err)
	}

	gossip_common.Log("Listening on ws://%s", addr)
	boot(listener)
}

func fetchName() {
	// Define the server name file path
	serverNameFile := filepath.Join(os.TempDir(), "gossip_server.name")

	// Check if the server name file exists
	if _, err := os.Stat(serverNameFile); os.IsNotExist(err) {
		// File does not exist, create it with a default server name
		defaultServerName := "Generic Gossip Server v" + version
		err := os.WriteFile(serverNameFile, []byte(defaultServerName+"\n"), 0644)
		if err != nil {
			gossip_common.Err("Failed to create server.name file: %v", err)
			return
		}
		serverName = defaultServerName
		gossip_common.Log("Server name file created and set to default: %s", serverName)
	} else {
		// File exists, read the server name from the file
		data, err := os.ReadFile(serverNameFile)
		if err != nil {
			gossip_common.Err("Failed to read server name from file: %v", err)
			return
		}

		// Extract the first line from the file content and limit to 64 characters
		firstLine := strings.Split(string(data), "\n")[0]
		if len(firstLine) > 64 {
			firstLine = firstLine[:64]
		}
		serverName = firstLine
		gossip_common.Log("Server name found and set to: %s", serverName)
	}
}

func compileChannels() {
	// Define the path for the channels file
	channelsFilePath := filepath.Join(os.TempDir(), "gossip_channels.list")

	// Check if the channels file exists
	if _, err := os.Stat(channelsFilePath); os.IsNotExist(err) {
		gossip_common.Log("Couldn't find channels.list, creating one with default channel 'general'...")
		// File does not exist, create it with "general" as the initial channel
		defaultChannel := []string{"general"}
		data := strings.Join(defaultChannel, "\n")
		err := os.WriteFile(channelsFilePath, []byte(data), 0644)
		if err != nil {
			gossip_common.Err("Failed to create channels.list file: %v", err)
		} else {
			channels = append(channels, defaultChannel...)
		}
	} else {
		// File exists, read the file and add strings to the channels slice
		gossip_common.Log("Found channels.list, importing...")
		data, err := os.ReadFile(channelsFilePath)
		if err != nil {
			gossip_common.Err("Failed to read channels.list file: %v", err)
		} else {
			channelNames := strings.Split(strings.TrimSpace(string(data)), "\n")
			channels = append(channels, channelNames...)
		}
	}
}
