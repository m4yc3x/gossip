package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"gossip_common"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	recordDevice    *Recorder
	playbackDevices = make(map[string]*Player)
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

/**
 * startup is called when the app starts. The context is saved
 * so we can call the runtime methods
 * @param ctx The context of the app
 */
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

/**
 * LoadSettings loads settings from a file
 * @return Settings Loaded settings
 * @return error Error if any occurred during loading
 */
func (a *App) LoadSettings() (Settings, error) {
	return LoadSettings()
}

/**
 * SaveSettings saves provided settings to a file
 * @param settings Settings to save
 * @return error Error if any occurred during saving
 */
func (a *App) SaveSettings(settings Settings) error {
	return SaveSettings(settings)
}

/**
 * LoadSettings loads settings from a file
 * @return Settings Loaded settings
 * @return error Error if any occurred during loading
 */
func LoadSettings() (Settings, error) {
	var settings Settings
	data, err := ioutil.ReadFile(filepath.Join(os.TempDir(), "gossip_settings.json"))
	if err != nil {
		if os.IsNotExist(err) {
			// If the file does not exist, create it with default settings
			defaultSettings := Settings{
				SelectedTheme:   "wintry",
				DefaultUsername: "",
				DefaultHost:     "",
				DefaultPort:     "1720",
			} // Assuming default settings are handled in the Settings struct
			data, _ := json.Marshal(defaultSettings)
			ioutil.WriteFile(filepath.Join(os.TempDir(), "gossip_settings.json"), data, 0644)
			return defaultSettings, nil
		}
		return settings, err
	}
	err = json.Unmarshal(data, &settings)
	return settings, err
}

/**
 * SaveSettings saves settings to a file
 * @param settings Settings to save
 * @return error Error if any occurred during saving
 */
func SaveSettings(settings Settings) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(os.TempDir(), "gossip_settings.json"), data, 0644)
}

/**
 * Boot initializes the connection and starts the main event loop
 * @param chost Host for the connection
 * @param cport Port for the connection
 * @param cusername Username for the connection
 * @param cpassword Password for the connection
 * @return error Error if any occurred during the boot process
 */
func (a *App) Boot(chost string, cport int, cusername string, cpassword string) error {

	host = chost
	port = cport
	username = cusername
	password = gossip_common.HashPassword(cpassword)

	runtime.EventsEmit(a.ctx, "update-loading-status", "Starting connection...")

	conn, writer = bootstrap(a) // Establish a connection (function not provided)
	if conn != nil {
		defer conn.Close() // Close the connection when the function returns

		runtime.EventsEmit(a.ctx, "update-loading-status", "Sending greeting to server...")

		go handleResponses(conn, writer, a) // Handle responses in a separate goroutine (function not provided)

		select {} // Block the main goroutine indefinitely
	}

	return nil
}

/**
 * SendMessage sends an encrypted message to all clients
 * @param message The message to send
 * @param expiry Expiry time of the message
 * @return error Error if any occurred during message sending
 */
func (a *App) SendMessage(message string, expiry int64, channel string) error {
	// Refresh the slice of public keys from the map
	publicKeysSlice := make([][]byte, 0, len(publicKeys))
	for _, publicKey := range publicKeys {
		publicKeysSlice = append(publicKeysSlice, publicKey)
	}

	// Encrypt the input value with all public keys
	encryptedMsg, err := gossip_common.GWEncryptToMultiple([]byte(message), publicKeysSlice)
	if err != nil {
		gossip_common.Err("Failed to encrypt PLD: %v", err)
		return nil
	}

	// Encrypt UID to all clients
	encryptedUID, err := gossip_common.GWEncryptToMultiple([]byte(username), publicKeysSlice)
	if err != nil {
		gossip_common.Err("Failed to encrypt UID: %v", err)
		return nil
	}

	// Create a data packet with the encrypted message
	cPacket := gossip_common.NewDataPacketFromData("cht", encryptedUID, time.Now().Unix(), expiry, 1, 1, gossip_common.GetClientID(), channel, encryptedMsg)

	// Send the data packet
	err = gossip_common.SendDataPacket(writer, cPacket, serverPublicKey)
	if err != nil {
		gossip_common.Err("Failed to send packet: %v", err)
		return nil
	}

	return nil
}

/**
 * Disconnect closes the current connection
 * @return error Error if any occurred during disconnection
 */
func (a *App) Disconnect() error {
	if recordDevice != nil {
		if recordDevice.device.IsStarted() {
			recordDevice.Stop()
		}
	}

	// Loop through all playback devices and stop each one if it's started
	for id, player := range playbackDevices {
		if player.device.IsStarted() {
			player.Stop()
		}
		delete(playbackDevices, id)
	}

	conn.Close()
	return nil
}

/**
 * SendAudioData sends audio data to the server
 * @param audioBlob The audio data to send
 * @return error Error if any occurred during audio sending
 */
func (a *App) StartRecording() {

	// Clear peers and data channels
	for key := range participentDataChannels {
		if dc := participentDataChannels[key]; dc != nil {
			dc.Close()
		}
		delete(participentDataChannels, key)
	}
	for key := range participentPeerConnections {
		if pc := participentPeerConnections[key]; pc != nil {
			pc.Close()
		}
		delete(participentPeerConnections, key)
	}

	recordDevice = NewRecorder()
	if err := recordDevice.Start(); err != nil {
		gossip_common.Err("Error starting recorder: %v", err)
		return
	}

	inCall = true

	if callID == "" {
		// Generate a random call ID
		callID = gossip_common.ReturnRandomID(24) // Assuming a function to generate random IDs

		runtime.EventsEmit(a.ctx, "call_starting")

		// Create a signal packet to start a call
		startCallPacket := gossip_common.NewSignalPacketFromData("start_call", "", gossip_common.GetClientID(), []byte(callID))

		// Send the start call packet
		err := gossip_common.SendSignalPacket(writer, startCallPacket)
		if err != nil {
			gossip_common.Err("Failed to send start call packet: %v", err)
			return
		}
	} else {
		GetParticipentsFromServer(callID, a)
	}
}

func (a *App) StopRecording() {

	// Clear peers and data channels
	for key := range participentDataChannels {
		if dc := participentDataChannels[key]; dc != nil {
			dc.Close()
		}
		delete(participentDataChannels, key)
	}
	for key := range participentPeerConnections {
		if pc := participentPeerConnections[key]; pc != nil {
			pc.Close()
		}
		delete(participentPeerConnections, key)
	}

	if recordDevice.device.IsStarted() {
		recordDevice.Stop()
	}

	// Loop through all playback devices and stop each one if it's started
	for id, player := range playbackDevices {
		if player.device.IsStarted() {
			player.Stop()
		}
		delete(playbackDevices, id)
	}

	runtime.EventsEmit(a.ctx, "caller_self_hung_up")

	HangUp(a)
	inCall = false
}

/**
 * UpdateCallID updates the call ID
 * @param callerID The call ID to update
 */
func (a *App) UpdateCallID(callerID string) {
	callID = callerID
}

/**
 * ToggleGoMute toggles the mute state of the client
 * @return error Error if any occurred during mute toggling
 */
func (a *App) ToggleGoMute() {
	muted = !muted
}

func (a *App) ToggleGoDeaf() {
	deafened = !deafened
}

func (a *App) cleanup(ctx context.Context) bool {
	if recordDevice.device.IsStarted() {
		recordDevice.Stop()
	}

	// Loop through all playback devices and stop each one if it's started
	for id, player := range playbackDevices {
		if player.device.IsStarted() {
			player.Stop()
		}
		delete(playbackDevices, id)
	}

	conn.Close()
	os.Exit(0)
	return true
}
