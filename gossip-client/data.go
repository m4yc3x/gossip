package main

import (
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"gossip_common"
)

type ChatMessage struct {
	Username   string `json:"username"`
	UID        string `json:"uid"`
	Message    string `json:"message"`
	Expiration int64  `json:"expiration"`
	Timestamp  int64  `json:"timestamp"`
}

/**
 * Handles incoming messages from the data channel.
 * @param msg The message received from the data channel.
 */
func handleDataPacket(packet gossip_common.GMDataPacket, a *App) {

	switch packet.OpCmd {
	case "cht":

		decryptedMsg, err := gossip_common.GWDecrypt(packet.Payload) // Decrypt the message
		if err != nil {
			gossip_common.Err("Failed to decrypt PLD: %v", err)
			break
		}

		decryptedUID, err := gossip_common.GWDecrypt(packet.UID) // Decrypt the message
		if err != nil {
			gossip_common.Err("Failed to decrypt UID: %v", err)
			break
		}

		formattedMsg := string(decryptedMsg) // Format the message for display

		if debugLogging {
			gossip_common.Dbg("CHT from %s", packet.Sender)
		}

		// send update to UI
		runtime.EventsEmit(a.ctx, "message-received", packet.Destination, string(decryptedUID), formattedMsg, packet.Expiration, packet.Timestamp, packet.Sender)
	}
}
