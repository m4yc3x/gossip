package main

import (
	"bufio"
	"encoding/json"
	"fmt"

	"gossip_common"

	"github.com/pion/webrtc/v4"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var participentPeerConnections = make(map[string]*webrtc.PeerConnection)
var participentDataChannels = make(map[string]*webrtc.DataChannel)

func GetParticipentsFromServer(callID string, a *App) { // only called by the joiner
	runtime.EventsEmit(a.ctx, "call_starting")

	// Create a signal packet to start a call
	getParticipentsPacket := gossip_common.NewSignalPacketFromData("gmp", "", gossip_common.GetClientID(), []byte(callID))

	// Send the start call packet
	err := gossip_common.SendSignalPacket(writer, getParticipentsPacket)
	if err != nil {
		gossip_common.Err("Failed to send start call packet: %v", err)
		return
	}

	if debugLogging {
		gossip_common.Dbg("GetParticipentsFromServer called")
	}
}

/**
 * SendOfferToClient initializes a WebRTC offer and sends it to the signaling server.
 * @return error Potential error during the offer creation or sending process.
 */
func SendOfferToClient(destination string, a *App, bufferWriter *bufio.Writer) error {
	// Create a new PeerConnection
	var err error
	participentPeerConnections[destination], err = webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create peer connection: %w", err)
	}

	// Create a data channel
	participentDataChannels[destination], err = participentPeerConnections[destination].CreateDataChannel("data", nil)
	if err != nil {
		return fmt.Errorf("failed to create data channel: %w", err)
	}

	participentDataChannels[destination].OnOpen(func() {
		if debugLogging {
			gossip_common.Dbg("Data channel opened")
		}
		gossip_common.Log("You are successfully connected to %s!", destination)

		runtime.EventsEmit(a.ctx, "call_started")
		runtime.EventsEmit(a.ctx, "caller_self_active")
		runtime.EventsEmit(a.ctx, "caller_active", destination)

	})

	participentDataChannels[destination].OnClose(func() {
		if debugLogging {
			gossip_common.Dbg("Data channel closed")
		}
		runtime.EventsEmit(a.ctx, "caller_hung_up", destination)
	})

	participentDataChannels[destination].OnMessage(func(msg webrtc.DataChannelMessage) {

		if player, exists := playbackDevices[destination]; exists {
			player.AddToBuffer(msg.Data)
		} else {
			// Create a new player for the destination if it does not exist
			newPlayer := NewPlayer()
			if err := newPlayer.initDevice(make([][]byte, 0)); err != nil {
				return
			}
			if err := newPlayer.Start(); err != nil {
				gossip_common.Err("Failed to start new player for %s: %v", destination, err)
			} else {
				playbackDevices[destination] = newPlayer
				playbackDevices[destination].AddToBuffer(msg.Data)
			}
		}

	})

	participentPeerConnections[destination].OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			// ICE gathering is finished
			return
		}

		candidate := c.ToJSON()
		err := SendICECandidate(destination, candidate, bufferWriter)
		if err != nil {
			fmt.Printf("Failed to send ICE candidate: %v\n", err)
		}
	})

	// Create an offer
	offer, err := participentPeerConnections[destination].CreateOffer(nil)
	if err != nil {
		return fmt.Errorf("failed to create offer: %w", err)
	}

	// Set the local description
	err = participentPeerConnections[destination].SetLocalDescription(offer)
	if err != nil {
		return fmt.Errorf("failed to set local description: %w", err)
	}

	// Send the offer to the signaling server
	offerPayload, err := json.Marshal(offer)
	if err != nil {
		return fmt.Errorf("failed to marshal offer: %w", err)
	}

	offerPacket := gossip_common.NewSignalPacketFromData("offer", destination, gossip_common.GetClientID(), offerPayload)
	err = gossip_common.SendSignalPacket(bufferWriter, offerPacket)
	if err != nil {
		gossip_common.Err("Failed to send offer packet: %v", err)
		return fmt.Errorf("failed to send offer packet: %w", err)
	}

	if debugLogging {
		gossip_common.Dbg("Offer sent to %s", destination)
	}

	return nil
}

/**
 * HandleOffer processes an offer from the signaling server, creates an answer, and sends it back.
 * @param conn The connection to the signaling server.
 * @param offer The received offer to handle.
 * @return error Potential error during the answer creation or sending process.
 */
func HandleOffer(sender string, offerBlob []byte, bufferWriter *bufio.Writer, a *App) error {

	var offer webrtc.SessionDescription
	err := json.Unmarshal(offerBlob, &offer)
	if err != nil {
		return fmt.Errorf("failed to unmarshal offer: %v", err)
	}

	if participentPeerConnections[sender] == nil {
		var err error
		participentPeerConnections[sender], err = webrtc.NewPeerConnection(webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create peer connection: %w", err)
		}

		participentPeerConnections[sender].OnDataChannel(func(d *webrtc.DataChannel) {
			participentDataChannels[sender] = d
			d.OnOpen(func() {
				if debugLogging {
					gossip_common.Dbg("Data channel opened")
				}
				gossip_common.Log("You are successfully connected to %s!", sender)

				runtime.EventsEmit(a.ctx, "call_started")
				runtime.EventsEmit(a.ctx, "caller_self_active")
				runtime.EventsEmit(a.ctx, "caller_active", sender)
			})

			d.OnClose(func() {
				if debugLogging {
					gossip_common.Dbg("Data channel closed")
				}
				runtime.EventsEmit(a.ctx, "caller_hung_up", sender)
			})

			d.OnMessage(func(msg webrtc.DataChannelMessage) {
				// Process the received message
				// Process the received message
				if player, exists := playbackDevices[sender]; exists {
					player.AddToBuffer(msg.Data)
				} else {
					// Create a new player for the destination if it does not exist
					newPlayer := NewPlayer()
					if err := newPlayer.initDevice(make([][]byte, 0)); err != nil {
						return
					}
					if err := newPlayer.Start(); err != nil {
						gossip_common.Err("Failed to start new player for %s: %v", sender, err)
					} else {
						playbackDevices[sender] = newPlayer
						playbackDevices[sender].AddToBuffer(msg.Data)
					}
				}
			})
		})
	}

	// Set the remote description
	err = participentPeerConnections[sender].SetRemoteDescription(offer)
	if err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	participentPeerConnections[sender].OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			// ICE gathering is finished
			return
		}

		candidate := c.ToJSON()
		err := SendICECandidate(sender, candidate, bufferWriter)
		if err != nil {
			fmt.Printf("Failed to send ICE candidate: %v\n", err)
		}
	})

	// Create an answer
	answer, err := participentPeerConnections[sender].CreateAnswer(nil)
	if err != nil {
		return fmt.Errorf("failed to create answer: %w", err)
	}

	// Set the local description
	err = participentPeerConnections[sender].SetLocalDescription(answer)
	if err != nil {
		return fmt.Errorf("failed to set local description: %w", err)
	}

	// Send the answer to the signaling server
	answerPayload, err := json.Marshal(answer)
	if err != nil {
		return fmt.Errorf("failed to marshal answer: %w", err)
	}

	answerPacket := gossip_common.NewSignalPacketFromData("answer", sender, gossip_common.GetClientID(), answerPayload)
	err = gossip_common.SendSignalPacket(bufferWriter, answerPacket)
	if err != nil {
		gossip_common.Err("Failed to send answer packet: %v", err)
		return fmt.Errorf("failed to send answer packet: %w", err)
	}

	if debugLogging {
		gossip_common.Dbg("Answer sent to %s", sender)
	}

	return nil
}

/**
 * SendICECandidate sends an ICE candidate to the signaling server.
 * @param conn The connection to the signaling server.
 * @param candidate The ICE candidate to send.
 * @return error Potential error during the ICE candidate sending process.
 */
func SendICECandidate(destination string, candidate webrtc.ICECandidateInit, bufferWriter *bufio.Writer) error {
	candidatePayload, err := json.Marshal(candidate)
	if err != nil {
		return fmt.Errorf("failed to marshal ICE candidate: %w", err)
	}

	icePacket := gossip_common.NewSignalPacketFromData("ice", destination, gossip_common.GetClientID(), candidatePayload)
	err = gossip_common.SendSignalPacket(bufferWriter, icePacket)
	if err != nil {
		gossip_common.Err("Failed to send ice packet: %v", err)
		return fmt.Errorf("failed to send ice packet: %w", err)
	}

	if debugLogging {
		gossip_common.Dbg("ICE candidate sent to %s", destination)
	}

	return nil
}

/**
 * HandleICECandidate processes an ICE candidate received from the signaling server and adds it to the PeerConnection.
 * @param destination The destination client ID.
 * @param candidateBlob The ICE candidate data received.
 * @return error Potential error during the ICE candidate addition process.
 */
func HandleICECandidate(destination string, candidateBlob []byte) error {
	var candidate webrtc.ICECandidateInit
	err := json.Unmarshal(candidateBlob, &candidate)
	if err != nil {
		return fmt.Errorf("failed to unmarshal ICE candidate: %v", err)
	}

	// Check if the PeerConnection for the destination exists and is not nil
	pc, exists := participentPeerConnections[destination]
	if !exists || pc == nil {
		gossip_common.Err("peer connection does not exist or is nil for destination: %s", destination)
		return fmt.Errorf("peer connection does not exist or is nil for destination: %s", destination)
	}

	// Add the ICE candidate to the PeerConnection
	err = pc.AddICECandidate(candidate)
	if err != nil {
		gossip_common.Err("failed to add ICE candidate: %v", err)
		return fmt.Errorf("failed to add ICE candidate: %w", err)
	}

	if debugLogging {
		gossip_common.Dbg("ICE candidate added to %s", destination)
	}
	return nil
}

/**
 * HandleAnswer processes a received answer and sets it as the remote description.
 * @param answer The received answer to process.
 * @return error Potential error during the remote description setting process.
 */
func HandleAnswer(sender string, answerBlob []byte) error {
	var answer webrtc.SessionDescription
	err := json.Unmarshal(answerBlob, &answer)
	if err != nil {
		return fmt.Errorf("failed to unmarshal answer: %v", err)
	}

	// Set the remote description with the received answer
	err = participentPeerConnections[sender].SetRemoteDescription(answer)
	if err != nil {
		return fmt.Errorf("failed to set remote description: %v", err)
	}
	return nil
}

func HangUp(a *App) {
	// send hangup to all participants
	runtime.EventsEmit(a.ctx, "hang-up")

	// Create a signal packet to start a call
	hangupPacket := gossip_common.NewSignalPacketFromData("hang-up", "", gossip_common.GetClientID(), []byte(callID))

	// Send the start call packet
	err := gossip_common.SendSignalPacket(writer, hangupPacket)
	if err != nil {
		gossip_common.Err("Failed to send hangup packet: %v", err)
		return
	}

	if debugLogging {
		gossip_common.Dbg("Sent hangup packet")
	}

	// Stop all active peer connections, dispose of them, and close associated data channels
	for sender, pc := range participentPeerConnections {
		// Close and remove associated data channels
		if dc, exists := participentDataChannels[sender]; exists && dc != nil {
			dc.Close()
			delete(participentDataChannels, sender)
			if debugLogging {
				gossip_common.Dbg("Closed data channel for %s", sender)
			}
		}
		if pc != nil {
			// Close the peer connection
			err := pc.Close()
			if err != nil {
				gossip_common.Err("Failed to close peer connection for %s: %v", sender, err)
			} else {
				if debugLogging {
					gossip_common.Dbg("Closed peer connection for %s", sender)
				}
			}
		}
		// Remove the peer connection from the map
		delete(participentPeerConnections, sender)
	}
}

func SendAudioToChannels(pSample []byte) {
	// Iterate over all participant data channels
	for id, dc := range participentDataChannels {
		if dc.ReadyState() == webrtc.DataChannelStateOpen {

			// Check if the public key exists for the participant
			if publicKey, exists := publicKeys[id]; !exists || publicKey == nil {
				// Close the data channel if no public key is found
				dc.Close()
				delete(participentDataChannels, id)
				if debugLogging {
					gossip_common.Dbg("Closed data channel for %s due to missing public key", id)
				}
				continue
			}

			// Encrypt the audio sample with the recipient's PGP key
			encryptedSample, err := gossip_common.GWEncrypt(pSample, publicKeys[id])
			if err != nil {
				gossip_common.Err("Failed to encrypt audio for %s: %v", id, err)
				continue
			}

			// Send the audio sample if the data channel is open
			err = dc.Send(encryptedSample)
			if err != nil {
				gossip_common.Err("Failed to send audio to %s: %v", id, err)
			}
		}
	}
}
