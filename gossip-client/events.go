package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"

	"gossip_common"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	serverPublicKey []byte
)

/**
 * bootstrap establishes a connection to the signaling server
 * and sends a greeting packet.
 * @return conn The connection to the signaling server.
 * @return writer The buffered writer for the connection.
 */
func bootstrap(a *App) (net.Conn, *bufio.Writer) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		gossip_common.Err("Failed to connect to gossip server: %v", err)
		runtime.EventsEmit(a.ctx, "server-disconnect")
		return nil, nil
	}

	// Wrap the connection with a bufio.Writer
	writer := bufio.NewWriterSize(conn, (8192 * 2))

	// Send greeting packet
	greetingPacket := gossip_common.NewSignalPacketFromData("grtng", "", gossip_common.GetClientID(), gossip_common.RetrievePublicKey())

	err = gossip_common.SendSignalPacket(writer, greetingPacket)
	if err != nil {
		gossip_common.Err("Failed to send greeting packet: %v", err)
		return nil, nil
	}

	return conn, writer
}

/**
 * handleResponses reads and processes responses from the signaling server.
 * @param conn The connection to the signaling server.
 * @param writer The buffered writer for the connection.
 * @param a The application instance.
 */
func handleResponses(conn net.Conn, writer *bufio.Writer, a *App) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()

		if debugLogging {
			gossip_common.Dbg("RCVMSG: %s", message)
			gossip_common.Dbg("<sizeof>: %d bytes", len(message))
		}

		var packet *gossip_common.GMSigPacket
		var err error
		if message[0] == '0' {
			packet, err = gossip_common.DeserializeGMSigPacket(message[1:])
			if err != nil {
				gossip_common.Err("Failed to deserialize GMSigPacket: %v", err)
				continue
			}
			// Handle GMSigPacket (existing logic can be used here)
		} else if message[0] == '1' {
			// decode the packet
			decodedPacket, err := base64.StdEncoding.DecodeString(string(message[1:]))
			if err != nil {
				gossip_common.Err("Failed to decode base64 GMDataPacket: %v", err)
				continue
			}
			// decrypt the data packet
			decryptedPacket, err := gossip_common.GWDecrypt([]byte(decodedPacket))
			if err != nil {
				gossip_common.Err("Failed to decrypt GMDataPacket: %v", err)
				continue
			}

			dataPacket, err := gossip_common.DeserializeGMDataPacket(string(decryptedPacket))
			if err != nil {
				gossip_common.Err("Failed to deserialize GMDataPacket: %v", err)
				continue
			}
			handleDataPacket(*dataPacket, a)
			continue

		} else if message[0] == '2' {
			// decode the packet
			// decodedPacket, err := base64.StdEncoding.DecodeString(string(message[1:]))
			// if err != nil {
			// 	gossip_common.Err("Failed to decode base64 GMStreamPacket: %v", err)
			// 	continue
			// }
			// // decrypt the data packet
			// decryptedPacket, err := gossip_common.GWDecrypt([]byte(decodedPacket))
			// if err != nil {
			// 	gossip_common.Err("Failed to decrypt GMStreamPacket: %v", err)
			// 	continue
			// }

			// streamPacket, err := gossip_common.DeserializeGMStreamPacket(string(decryptedPacket))
			// if err != nil {
			// 	gossip_common.Err("Failed to deserialize GMStreamPacket: %v", err)
			// 	continue
			// }
			//HandleStreamPacket(*streamPacket, a)
			continue
		} else {
			gossip_common.Err("Unknown packet type prefix: %v", message[0])
			continue
		}

		// Handle received packet based on operation command
		switch packet.OpCmd {
		case "hru": // how are you packet

			serverPublicKey = packet.Payload
			if debugLogging {
				gossip_common.Dbg("Server's public key retrieved and stored.")
			}

			// Encrypt the message with the server's public key
			encryptedMsg, err := gossip_common.GWEncrypt([]byte(password), serverPublicKey)
			if err != nil {
				gossip_common.Err("Failed to encrypt message: %v", err)
				continue
			}

			serverName = packet.Sender

			// Create and send the encrypted message packet
			msgPacket := gossip_common.NewSignalPacketFromData("ighru", "", gossip_common.GetClientID(), encryptedMsg)
			err = gossip_common.SendSignalPacket(writer, msgPacket)
			if err != nil {
				gossip_common.Err("Failed to send encrypted message packet: %v", err)
				continue
			}

			runtime.EventsEmit(a.ctx, "update-loading-status", "Sending IGHRU to server...")
			runtime.EventsEmit(a.ctx, "server-name-received", serverName)
			runtime.EventsEmit(a.ctx, "update-client-id", gossip_common.GetClientID())

		case "ig": // i'm good packet
			// Decrypt the payload
			decryptedMsg, err := gossip_common.GWDecrypt(packet.Payload)
			if err != nil {
				gossip_common.Err("Failed to decrypt IG message: %v", err)
				continue
			}

			// Check if the decrypted message is "I'm good!"
			if string(decryptedMsg) != "I'm good!" {
				gossip_common.Err("Incorrect IG response")
				continue
			}

			if debugLogging {
				gossip_common.Dbg("Correct IG response received.")
			}

			gossip_common.Log("Securely connected to server!")

			// Create and send a "give me keys" packet
			msgPacket := gossip_common.NewSignalPacketFromData("gmk", "", gossip_common.GetClientID(), []byte(""))
			err = gossip_common.SendSignalPacket(writer, msgPacket)
			if err != nil {
				gossip_common.Err("Failed to send encrypted message packet: %v", err)
				continue
			}

			runtime.EventsEmit(a.ctx, "update-loading-status", "Sending GMK to server...")

			gossip_common.Log("Awaiting client keys...")

		case "ckp": // client key packet
			// Decrypt the public key from the payload
			decryptedKey, err := gossip_common.GWDecrypt(packet.Payload)
			if err != nil {
				gossip_common.Err("Failed to decrypt public key: %v", err)
				continue
			}

			// Store the decrypted public key in the publicKeys map
			publicKeys[packet.Sender] = decryptedKey

			runtime.EventsEmit(a.ctx, "update-loading-status", "Received key #"+string(len(publicKeys))+"...")

		case "cup": // channel update packet
			// Decrypt the channel from the payload
			decryptedChannel, err := gossip_common.GWDecrypt(packet.Payload)
			if err != nil {
				gossip_common.Err("Failed to decrypt public key: %v", err)
				continue
			}

			if debugLogging {
				gossip_common.Dbg("Received channel update")
			}

			runtime.EventsEmit(a.ctx, "channel-update", string(decryptedChannel))

		case "eok": // end of keys packet
			gossip_common.Log("Client keys received. Booting...")
			runtime.EventsEmit(a.ctx, "finish-loading-status")

		case "call_active":
			runtime.EventsEmit(a.ctx, "call_active", callID)

		case "participent":
			// send offer to the destination
			runtime.EventsEmit(a.ctx, "call_sending_offer")
			SendOfferToClient(packet.Destination, a, writer)

		case "offer":
			// Send the offer to the signaling server
			runtime.EventsEmit(a.ctx, "call_received_offer")
			HandleOffer(packet.Sender, packet.Payload, writer, a)

		case "answer":
			runtime.EventsEmit(a.ctx, "call_received_answer")
			HandleAnswer(packet.Sender, packet.Payload)

		case "ice":
			runtime.EventsEmit(a.ctx, "call_received_ice", callID)
			HandleICECandidate(packet.Sender, packet.Payload)

		case "c404": // call not found packet
			runtime.EventsEmit(a.ctx, "call_not_found", callID)

		case "401": // forbidden packet
			gossip_common.Err("Wrong server key!")
			conn.Close()
			if debugLogging {
				gossip_common.Dbg("Connection closed due to unauthorized access.")
			}

			runtime.EventsEmit(a.ctx, "unauthorized")
			continue

		case "rmk": // remove key packet
			if debugLogging {
				gossip_common.Dbg("RMK received from %s. Removing their key.", packet.Sender)
			}

			// Remove the key from the publicKeys map
			delete(publicKeys, packet.Sender)

		default:
			gossip_common.Err("Unknown operation command: %s", packet.OpCmd)

		}
	}

	if err := scanner.Err(); err != nil {
		gossip_common.Err("Error reading from connection: %v", err)
	}
}
