package main

import (
	"bufio"
	"encoding/base64"
	"net"
	"sync"

	"gossip_common"
)

var activeCalls = make(map[string][]string)
var activeCallsLock = sync.RWMutex{}

// boot listens for incoming connections and processes them.
func boot(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			gossip_common.Err("Failed to accept connection: %v", err)
			continue
		}

		// Handle each connection in a new goroutine.
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	writer := bufio.NewWriterSize(conn, (8192 * 2))

	if connectionLogging {
		gossip_common.Conn("Connection established from %v", conn.RemoteAddr())
	}

	scanner := bufio.NewScanner(conn)
	clientID := ""
	var clientPublicKey []byte

	defer func() {
		if clientID != "" {
			connectionsLock.Lock()
			delete(connections, clientID)
			delete(publicKeys, clientID)
			connectionsLock.Unlock()
			if connectionLogging {
				gossip_common.Conn("Client %s unregistered due to connection termination", clientID)
			}
			sendRMKPackets(clientID)
		}
	}()

	for scanner.Scan() {
		message := scanner.Text()

		if debugLogging {
			gossip_common.Dbg("RCVMSG: %s", message)
		}

		var packet *gossip_common.GMSigPacket
		var err error
		if message[0] == '0' {
			packet, err = gossip_common.DeserializeGMSigPacket(message[1:])
			if err != nil {
				gossip_common.Err("Failed to deserialize GMSigPacket: %v", err)
				continue
			}
		} else if message[0] == '1' {
			go handleDataPacket([]byte(message[1:]))
			continue
		} else if message[0] == '2' {
			go handleStreamPacket([]byte(message[1:]))
			continue
		} else {
			gossip_common.Err("Unknown packet type prefix: %v", message[0])
			continue
		}

		switch packet.OpCmd {

		case "grtng":

			// Register client connection
			clientID = packet.Sender
			clientPublicKey = packet.Payload

			connectionsLock.Lock()
			connections[clientID] = conn
			publicKeys[clientID] = clientPublicKey
			connectionsLock.Unlock()
			if connectionLogging {
				gossip_common.Conn("Client %s registered with public key", clientID)
			}

			sendKeyToOtherClients(clientID, clientPublicKey)

			// Send HRU
			responsePacket := gossip_common.NewSignalPacketFromData("hru", clientID, serverName, gossip_common.RetrievePublicKey())
			if err := gossip_common.SendSignalPacket(writer, responsePacket); err != nil {
				gossip_common.Err("Failed to send HRU packet to %s: %v", clientID, err)
				continue
			}

			if debugLogging {
				gossip_common.Dbg("Sent HRU response to %s", clientID)
			}

		case "ighru":
			// Decrypt the payload
			decryptedMsg, err := gossip_common.GWDecrypt(packet.Payload)
			if err != nil {
				gossip_common.Err("Failed to decrypt IGHRU message from %s: %v", clientID, err)
				continue
			}

			clientPassword := string(decryptedMsg)

			if clientPassword != password {
				gossip_common.Err("Failed to authenticate client %s", clientID)

				// Send 401
				responsePacket := gossip_common.NewSignalPacketFromData("401", clientID, "", []byte(""))
				if err := gossip_common.SendSignalPacket(writer, responsePacket); err != nil {
					gossip_common.Err("Failed to send 401 packet to %s: %v", clientID, err)
					continue
				}

				if debugLogging {
					gossip_common.Dbg("Sent 401 response to %s", clientID)
				}

				continue
			}

			if debugLogging {
				gossip_common.Dbg("Correct IGHRU response received from %s", clientID)
			}

			// Encrypt the payload "I'm good!" with the client's public key
			encryptedMsg, err := gossip_common.GWEncrypt([]byte("I'm good!"), publicKeys[clientID])
			if err != nil {
				gossip_common.Err("Failed to encrypt IG message for %s: %v", clientID, err)
				continue
			}

			// Create a new data packet with "ig" opcmd and the encrypted message as payload
			responsePacket := gossip_common.NewSignalPacketFromData("ig", clientID, "", encryptedMsg)

			// Send the encrypted "I'm good!" message back to the client
			if err := gossip_common.SendSignalPacket(writer, responsePacket); err != nil {
				gossip_common.Err("Failed to send IG packet to %s: %v", clientID, err)
				continue
			}

			if debugLogging {
				gossip_common.Dbg("Sent IG response to %s", clientID)
			}

		case "gmk":
			clientID = packet.Sender

			// Then, send all other clients' public keys to the requesting client
			for id, key := range publicKeys {
				// Encrypt each public key with the requesting client's public key
				encryptedKey, err := gossip_common.GWEncrypt(key, publicKeys[clientID])
				if err != nil {
					gossip_common.Err("Failed to encrypt public key for %s: %v", clientID, err)
					continue
				}

				// Create a new data packet with "ckp" opcmd (client key packet) and the encrypted key as payload
				keyPacket := gossip_common.NewSignalPacketFromData("ckp", clientID, string(id), encryptedKey)

				// Send the encrypted public key to the client
				if err := gossip_common.SendSignalPacket(writer, keyPacket); err != nil {
					gossip_common.Err("Failed to send key packet to %s: %v", clientID, err)
					continue
				}

				if debugLogging {
					gossip_common.Dbg("Sent encrypted public key to %s", clientID)
				}
			}

			// Loop through the channels and send a "cup" (channel update) packet with the channel name encrypted with the client's public key
			for _, channel := range channels {
				// Encrypt the channel name with the requesting client's public key
				encryptedChannelName, err := gossip_common.GWEncrypt([]byte(channel), publicKeys[clientID])
				if err != nil {
					gossip_common.Err("Failed to encrypt channel name for %s: %v", clientID, err)
					continue
				}

				// Create a new data packet with "cup" opcmd (channel update packet) and the encrypted channel name as payload
				channelPacket := gossip_common.NewSignalPacketFromData("cup", clientID, "", encryptedChannelName)

				// Send the encrypted channel name to the client
				if err := gossip_common.SendSignalPacket(writer, channelPacket); err != nil {
					gossip_common.Err("Failed to send channel packet to %s: %v", clientID, err)
					continue
				}

				if debugLogging {
					gossip_common.Dbg("Sent encrypted channel %s to %s", channel, clientID)
				}
			}

			// Once all keys have been sent, send an "eok" (end of keys) signal packet to the requesting client
			eokPacket := gossip_common.NewSignalPacketFromData("eok", clientID, "", []byte(""))
			if err := gossip_common.SendSignalPacket(writer, eokPacket); err != nil {
				gossip_common.Err("Failed to send EOK packet to %s: %v", clientID, err)
				continue
			}
			if debugLogging {
				gossip_common.Dbg("Sent EOK to %s", clientID)
			}
		case "start_call":
			activeCallsLock.Lock()
			activeCalls[string(packet.Payload)] = append(activeCalls[string(packet.Payload)], packet.Sender)
			activeCallsLock.Unlock()

			csPacket := gossip_common.NewSignalPacketFromData("call_active", clientID, "", []byte(""))
			if err := gossip_common.SendSignalPacket(writer, csPacket); err != nil {
				gossip_common.Err("Failed to send call_active packet to %s: %v", clientID, err)
				continue
			}
			if debugLogging {
				gossip_common.Dbg("Sent call_active to %s", clientID)
			}

		case "gmp": // give me participents
			// Check if the call ID exists in activeCalls
			if _, ok := activeCalls[string(packet.Payload)]; !ok {
				// If the call ID doesn't exist, send a "c404" (call not found) signal packet to the requesting client
				c404Packet := gossip_common.NewSignalPacketFromData("c404", clientID, "", []byte(""))
				if err := gossip_common.SendSignalPacket(writer, c404Packet); err != nil {
					gossip_common.Err("Failed to send c404 packet to %s: %v", clientID, err)
					continue
				}
				if debugLogging {
					gossip_common.Dbg("Sent c404 to %s", clientID)
				}
			} else {
				activeCalls[string(packet.Payload)] = append(activeCalls[string(packet.Payload)], packet.Sender)
				// Send each participant in its own packet
				for _, participant := range activeCalls[string(packet.Payload)] {
					if participant != packet.Sender {
						participantPacket := gossip_common.NewSignalPacketFromData("participent", participant, packet.Sender, []byte(participant))
						if err := gossip_common.SendSignalPacket(writer, participantPacket); err != nil {
							gossip_common.Err("Failed to send participant packet to %s: %v", clientID, err)
							continue
						}
						if debugLogging {
							gossip_common.Dbg("Sent participant %s to %s", participant, clientID)
						}
					}
				}
			}

		case "offer":

			destination := packet.Destination
			offerWriter := bufio.NewWriter(connections[destination])
			offer := packet.Payload

			offerPacket := gossip_common.NewSignalPacketFromData("offer", destination, packet.Sender, offer)
			if err := gossip_common.SendSignalPacket(offerWriter, offerPacket); err != nil {
				gossip_common.Err("Failed to send offer packet to %s: %v", destination, err)
				continue
			}
			if debugLogging {
				gossip_common.Dbg("Relayed offer from %s to %s", packet.Sender, destination)
			}

		case "answer":
			destination := packet.Destination
			answerWriter := bufio.NewWriter(connections[destination])
			answer := packet.Payload

			answerPacket := gossip_common.NewSignalPacketFromData("answer", destination, packet.Sender, answer)
			if err := gossip_common.SendSignalPacket(answerWriter, answerPacket); err != nil {
				gossip_common.Err("Failed to send answer packet to %s: %v", destination, err)
				continue
			}
			if debugLogging {
				gossip_common.Dbg("Relayed answer from %s to %s", packet.Sender, destination)
			}

		case "ice":
			destination := packet.Destination
			iceWriter := bufio.NewWriter(connections[destination])
			candidate := packet.Payload

			icePacket := gossip_common.NewSignalPacketFromData("ice", destination, packet.Sender, candidate)
			if err := gossip_common.SendSignalPacket(iceWriter, icePacket); err != nil {
				gossip_common.Err("Failed to send ice packet to %s: %v", destination, err)
				continue
			}
			if debugLogging {
				gossip_common.Dbg("Relayed ice from %s to %s", packet.Sender, destination)
			}

		case "hang-up":
			/**
			 * Handle hang-up signal by removing the sender from the call participants list.
			 * @param packet The received signal packet containing the call ID in the payload.
			 */
			callID := string(packet.Payload)
			sender := packet.Sender

			// Lock the calls map to safely update
			activeCallsLock.Lock()
			if participants, ok := activeCalls[callID]; ok {
				// Remove the sender from the call participants list
				newParticipants := []string{}
				for _, participant := range participants {
					if participant != sender {
						newParticipants = append(newParticipants, participant)
					}
				}
				// Update the call participants in the map
				if len(newParticipants) > 0 {
					activeCalls[callID] = newParticipants
				} else {
					// If no participants left, delete the call entry
					delete(activeCalls, callID)
				}
			}
			activeCallsLock.Unlock()

			if debugLogging {
				gossip_common.Dbg("Handled hang-up from %s for call %s", sender, callID)
			}

		case "urgstr":

			if clientID != "" {
				connectionsLock.Lock()
				delete(connections, clientID)
				delete(publicKeys, clientID)
				connectionsLock.Unlock()
				if connectionLogging {
					gossip_common.Conn("Client %s unregistered", clientID)
				}
				sendRMKPackets(clientID)
			}

		}

	}

	if debugLogging {
		if err := scanner.Err(); err != nil {
			gossip_common.Err("Error reading from connection: %v", err)
		}
	}
}

/**
 * handleDataPacket accepts a message, deserializes it, and forwards it to all recipients.
 * @param message The message to be forwarded.
 */
func handleDataPacket(message []byte) {

	// Decode the base64 packet
	decodedPacket, err := base64.StdEncoding.DecodeString(string(message))
	if err != nil {
		if debugLogging {
			gossip_common.Err("Failed to decode base64 GMDataPacket: %v", err)
		}
		return
	}

	// decrypt the packet before deserializing
	decryptedPacket, err := gossip_common.GWDecrypt(decodedPacket)
	if err != nil && debugLogging {
		gossip_common.Err("Failed to decrypt GMDataPacket: %v", err)
		return
	}

	// Deserialize the message into a GMDataPacket
	dataPacket, err := gossip_common.DeserializeGMDataPacket(string(decryptedPacket))
	if err != nil {
		gossip_common.Err("Failed to deserialize message: %v", err)
		return
	}

	for id, conn := range connections {
		writer := bufio.NewWriter(conn)
		// Reserialize and send the packet using SendDataPacket
		if err := gossip_common.SendDataPacket(writer, *dataPacket, publicKeys[id]); err != nil {
			if debugLogging {
				gossip_common.Err("Failed to forward message: %v", err)
			}
			continue
		}
	}

	if debugLogging {
		gossip_common.Dbg("Message forwarded to all clients")
	}
}

/**
 * handleStreamPacket accepts a message, deserializes it, and forwards it to specific recipients noted in the packet.
 * @param message The message to be forwarded.
 */
func handleStreamPacket(message []byte) {
	// Decode the base64 packet
	decodedPacket, err := base64.StdEncoding.DecodeString(string(message))
	if err != nil {
		if debugLogging {
			gossip_common.Err("Failed to decode base64 GMStreamPacket: %v", err)
		}
		return
	}

	// Decrypt the packet before deserializing
	decryptedPacket, err := gossip_common.GWDecrypt(decodedPacket)
	if err != nil {
		if debugLogging {
			gossip_common.Err("Failed to decrypt GMStreamPacket: %v", err)
		}
		return
	}

	// Deserialize the message into a GMStreamPacket
	streamPacket, err := gossip_common.DeserializeGMStreamPacket(string(decryptedPacket))
	if err != nil {
		if debugLogging {
			gossip_common.Err("Failed to deserialize GMStreamPacket: %v", err)
		}
		return
	}

	// Forward the packet only to the recipients listed in the packet
	for _, recipientID := range streamPacket.Recipients {
		conn, exists := connections[recipientID]
		if !exists {
			continue
		}

		writer := bufio.NewWriter(conn)
		// Reserialize and send the packet using SendDataPacket
		if err := gossip_common.SendStreamPacket(writer, *streamPacket, publicKeys[recipientID]); err != nil {
			continue
		}
	}

	if debugLogging {
		gossip_common.Dbg("Stream message forwarded to specified clients")
	}
}

func sendRMKPackets(clientID string) {
	// Notify all clients that a client has unregistered
	for id, conn := range connections {
		writer := bufio.NewWriter(conn)
		if id != clientID {
			// Create a "rmk" (remove key) signal packet to notify other clients
			rmkPacket := gossip_common.NewSignalPacketFromData("rmk", id, clientID, []byte(""))
			if err := gossip_common.SendSignalPacket(writer, rmkPacket); err != nil {
				gossip_common.Err("Failed to send 'rmk' packet to %s: %v", id, err)
				continue
			}
			if debugLogging {
				gossip_common.Dbg("Sent 'rmk' notification about %s to %s", clientID, id)
			}
		}
	}
}

/**
 * sendClientKeyToOtherClients sends the public key of a client to all other connected clients except itself.
 * @param clientID The ID of the client whose key is to be sent.
 * @param clientPublicKey The public key of the client to be sent.
 */
func sendKeyToOtherClients(clientID string, clientPublicKey []byte) {
	connectionsLock.RLock()
	defer connectionsLock.RUnlock()

	for id, conn := range connections {

		writer := bufio.NewWriter(conn)
		// Encrypt the public key with the recipient client's public key before sending
		recipientPublicKey, exists := publicKeys[id]
		if !exists {
			gossip_common.Err("No public key found for client %s, cannot encrypt.", id)
			continue
		}
		encryptedPublicKey, err := gossip_common.GWEncrypt(clientPublicKey, recipientPublicKey)
		if err != nil {
			gossip_common.Err("Failed to encrypt client key for %s: %v", id, err)
			continue
		}

		// Create a "ckp" (client key packet) signal packet to send the encrypted public key
		ckpPacket := gossip_common.NewSignalPacketFromData("ckp", id, clientID, encryptedPublicKey)
		if err := gossip_common.SendSignalPacket(writer, ckpPacket); err != nil {
			gossip_common.Err("Failed to send encrypted client key to %s: %v", id, err)
			continue
		}
		if debugLogging {
			gossip_common.Dbg("Sent %s key to %s", clientID, id)
		}
	}
}
