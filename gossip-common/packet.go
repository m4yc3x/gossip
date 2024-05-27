package gossip_common

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

/**
 * GMSigPacket defines the structure for signaling messages between peers.
 * @param OpCmd The operation command to determine the action.
 * @param Destination The intended recipient of the message.
 * @param Sender The sender of the message.
 * @param Payload The data being sent.
 */
type GMSigPacket struct {
	OpCmd       string `json:"cmd"` // Operation command to determine the action
	Destination string `json:"dst"` // The intended recipient of the message
	Sender      string `json:"snd"` // The sender of the message
	Payload     []byte `json:"pld"` // The data being sent
}

/**
 * GMDataPacket defines the structure for data messages in the gowire application.
 * It includes metadata for message handling and fragmentation.
 *
 * @param OpCmd Operation command to determine the action required upon receiving this packet.
 * @param UID Unique identifier for the sender, ensuring message traceability and integrity.
 * @param Timestamp The creation time of the message, used for logging and possibly message ordering.
 * @param Expiration The time at which the message should be considered expired and no longer valid.
 * @param ChunkIndex The index of the current chunk in a sequence of fragmented message parts.
 * @param ChunkMax The total number of chunks that the original message has been divided into.
 * @param Sender The sender's identifier, providing context for the recipient.
 * @param Payload The actual data being sent, encapsulated in this packet structure.
 */
type GMDataPacket struct {
	OpCmd       string `json:"cmd"` // Operation command to determine the action
	UID         []byte `json:"uid"` // Unique identifier for the sender
	Timestamp   int64  `json:"ts"`  // Timestamp for the message
	Expiration  int64  `json:"exp"` // Expiration time for the message
	ChunkIndex  int64  `json:"idx"` // Index of the chunk
	ChunkMax    int64  `json:"max"` // Maximum number of chunks
	Sender      string `json:"snd"` // The sender of the message
	Destination string `json:"dst"` // The intended recipient of the message
	Payload     []byte `json:"pld"` // The data being sent
}

/**
 * GMStreamPacket defines the structure for streaming data messages such as voice and video.
 * It includes a list of recipients to allow the server to relay the packet to specific clients.
 * @param GMDataPacket The data packet to be sent.
 * @param Recipients List of recipient client IDs.
 */
type GMStreamPacket struct {
	GMDataPacket          // Embedding GMDataPacket to reuse existing fields
	Recipients   []string `json:"r"` // List of recipient client IDs
}

/**
 * NewMessagePacket creates a new instance of a GMSigPacket with string data.
 * @param OpCmd The operation command.
 * @param Destination The intended recipient's ID.
 * @param Sender The sender's ID.
 * @param message The message to be sent.
 * @return A new instance of GMSigPacket.
 */
func NewSignalPacketFromString(OpCmd string, Destination string, Sender string, message string) GMSigPacket {
	return GMSigPacket{
		OpCmd:       OpCmd,
		Destination: Destination,
		Sender:      Sender,
		Payload:     []byte(message),
	}
}

/**
 * NewDataPacket creates a new instance of a GMSigPacket with byte data.
 * @param OpCmd The operation command.
 * @param Destination The intended recipient's ID.
 * @param Sender The sender's ID.
 * @param data The data to be sent.
 * @return A new instance of GMSigPacket.
 */
func NewSignalPacketFromData(OpCmd string, Destination string, Sender string, data []byte) GMSigPacket {
	return GMSigPacket{
		OpCmd:       OpCmd,
		Destination: Destination,
		Sender:      Sender,
		Payload:     data,
	}
}

/**
 * NewDataPacketFromString creates a new instance of a GMDataPacket with string data.
 * @param OpCmd The operation command.
 * @param UID The unique identifier for the sender.
 * @param Timestamp The timestamp for the message.
 * @param Expiration The expiration time for the message.
 * @param ChunkIndex The index of the current chunk in a sequence of fragmented message parts.
 * @param ChunkMax The total number of chunks that the original message has been divided into.
 * @param Sender The sender's identifier.
 * @param message The message to be sent.
 * @return A new instance of GMDataPacket.
 */
func NewDataPacketFromString(OpCmd string, UID string, Timestamp int64, Expiration int64, ChunkIndex int64, ChunkMax int64, Sender string, Destination string, message string) GMDataPacket {
	return GMDataPacket{
		OpCmd:       OpCmd,
		UID:         []byte(UID),
		Timestamp:   Timestamp,
		Expiration:  Expiration,
		ChunkIndex:  ChunkIndex,
		ChunkMax:    ChunkMax,
		Sender:      Sender,
		Destination: Destination,
		Payload:     []byte(message),
	}
}

/**
 * NewDataPacketFromData creates a new instance of a GMDataPacket with byte data.
 * @param OpCmd The operation command.
 * @param UID The unique identifier for the sender.
 * @param Timestamp The timestamp for the message.
 * @param Expiration The expiration time for the message.
 * @param ChunkIndex The index of the current chunk in a sequence of fragmented message parts.
 * @param ChunkMax The total number of chunks that the original message has been divided into.
 * @param Sender The sender's identifier.
 * @param data The data to be sent.
 * @return A new instance of GMDataPacket.
 */
func NewDataPacketFromData(OpCmd string, UID []byte, Timestamp int64, Expiration int64, ChunkIndex int64, ChunkMax int64, Sender string, Destination string, data []byte) GMDataPacket {
	return GMDataPacket{
		OpCmd:       OpCmd,
		UID:         UID,
		Timestamp:   Timestamp,
		Expiration:  Expiration,
		ChunkIndex:  ChunkIndex,
		ChunkMax:    ChunkMax,
		Sender:      Sender,
		Destination: Destination,
		Payload:     data,
	}
}

/**
 * NewStreamPacketFromData creates a new instance of a GMStreamPacket with byte data.
 * @param OpCmd The operation command.
 * @param UID The unique identifier for the sender.
 * @param Timestamp The timestamp for the message.
 * @param Expiration The expiration time for the message.
 * @param ChunkIndex The index of the current chunk in a sequence of fragmented message parts.
 * @param ChunkMax The total number of chunks that the original message has been divided into.
 * @param Sender The sender's identifier.
 * @param data The data to be sent.
 * @param recipients The list of recipient client IDs.
 * @return A new instance of GMStreamPacket.
 */
func NewStreamPacketFromData(OpCmd string, UID []byte, Timestamp int64, Expiration int64, ChunkIndex int64, ChunkMax int64, Sender string, Destination string, data []byte, recipients []string) GMStreamPacket {
	return GMStreamPacket{
		GMDataPacket: NewDataPacketFromData(OpCmd, UID, Timestamp, Expiration, ChunkIndex, ChunkMax, Sender, Destination, data),
		Recipients:   recipients,
	}
}

/**
 * SerializeGMSigPacket serializes a GMSigPacket into a JSON string.
 * @param packet The GMSigPacket to serialize.
 * @return The serialized JSON string and an error if any.
 */
func SerializeGMSigPacket(packet *GMSigPacket) (string, error) {
	packetBytes, err := json.Marshal(packet)
	if err != nil {
		return "", fmt.Errorf("failed to serialize GMSigPacket: %w", err)
	}
	return string(packetBytes), nil
}

/**
 * DeserializeGMSigPacket deserializes a JSON string into a GMSigPacket.
 * @param packetStr The JSON string to deserialize.
 * @return The deserialized GMSigPacket and an error if any.
 */
func DeserializeGMSigPacket(packetStr string) (*GMSigPacket, error) {
	var packet GMSigPacket
	err := json.Unmarshal([]byte(packetStr), &packet)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize GMSigPacket: %w", err)
	}
	return &packet, nil
}

/**
 * SerializeGMDataPacket serializes a GMDataPacket into a JSON string.
 * @param packet The GMDataPacket to serialize.
 * @return The serialized JSON string and an error if any.
 */
func SerializeGMDataPacket(packet *GMDataPacket) (string, error) {
	packetBytes, err := json.Marshal(packet)
	if err != nil {
		return "", fmt.Errorf("failed to serialize GMDataPacket: %w", err)
	}
	return string(packetBytes), nil
}

/**
 * DeserializeGMDataPacket deserializes a JSON string into a GMDataPacket.
 * @param packetStr The JSON string to deserialize.
 * @return The deserialized GMDataPacket and an error if any.
 */
func DeserializeGMDataPacket(packetStr string) (*GMDataPacket, error) {
	var packet GMDataPacket
	err := json.Unmarshal([]byte(packetStr), &packet)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize GMDataPacket: %w", err)
	}
	return &packet, nil
}

/**
 * SerializeGMStreamPacket serializes a GMStreamPacket into a JSON string.
 */
func SerializeGMStreamPacket(packet *GMStreamPacket) (string, error) {
	packetBytes, err := json.Marshal(packet)
	if err != nil {
		return "", fmt.Errorf("failed to serialize GMStreamPacket: %w", err)
	}
	return string(packetBytes), nil
}

/**
 * DeserializeGMStreamPacket deserializes a JSON string into a GMStreamPacket.
 */
func DeserializeGMStreamPacket(packetStr string) (*GMStreamPacket, error) {
	var packet GMStreamPacket
	err := json.Unmarshal([]byte(packetStr), &packet)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize GMStreamPacket: %w", err)
	}
	return &packet, nil
}

/**
 * Sends a GMSignalPacket over the given connection.
 * @param conn The connection to send the packet over.
 * @param packet The GMSignalPacket to be sent.
 * @return error An error if sending fails, nil otherwise.
 */
func SendSignalPacket(writer *bufio.Writer, packet GMSigPacket) error {
	packetStr, err := SerializeGMSigPacket(&packet)
	if err != nil {
		Err("Failed to serialize GMSigPacket: %v", err)
		return err
	}

	_, err = fmt.Fprintf(writer, "0%s\n", packetStr)
	if err != nil {
		Err("Failed to send GMSigPacket over connection: %v", err)
		return err
	}

	return writer.Flush()
}

/**
 * Sends a GMDataPacket over the given connection.
 * @param conn The connection to send the packet over.
 * @param packet The GMDataPacket to be sent.
 * @return error An error if sending fails, nil otherwise.
 */
func SendDataPacket(writer *bufio.Writer, packet GMDataPacket, publicKey []byte) error {
	packetStr, err := SerializeGMDataPacket(&packet)
	if err != nil {
		Err("Failed to serialize GMDataPacket: %v", err)
		return err
	}

	// Encrypt the serialized packet with the server's public key
	encryptedPacket, err := GWEncrypt([]byte(packetStr), publicKey)
	if err != nil {
		Err("Failed to encrypt GMDataPacket: %v", err)
		return err
	}

	// Convert the encrypted packet to a base64 string for transmission
	base64EncryptedPacket := base64.StdEncoding.EncodeToString(encryptedPacket)

	_, err = fmt.Fprintf(writer, "1%s\n", base64EncryptedPacket)
	if err != nil {
		Err("Failed to send GMDataPacket over connection: %v", err)
		return err
	}

	return writer.Flush()
}

/**
 * Sends a GMStreamPacket over the given connection, encrypting for each recipient.
 * @param writer The connection to send the packet over.
 * @param packet The GMStreamPacket to be sent.
 * @param serverPublicKey The public key of the server used to encrypt the packet.
 * @return error An error if sending fails, nil otherwise.
 */
func SendStreamPacket(writer *bufio.Writer, packet GMStreamPacket, serverPublicKey []byte) error {
	packetStr, err := SerializeGMStreamPacket(&packet)
	if err != nil {
		Err("Failed to serialize GMStreamPacket: %v", err)
		return err
	}

	// Encrypt the serialized packet with the server's public key
	encryptedPacket, err := GWEncrypt([]byte(packetStr), serverPublicKey)
	if err != nil {
		Err("Failed to encrypt GMStreamPacket: %v", err)
		return err
	}

	// Convert the encrypted packet to a base64 string for transmission
	base64EncryptedPacket := base64.StdEncoding.EncodeToString(encryptedPacket)

	_, err = fmt.Fprintf(writer, "2%s\n", base64EncryptedPacket)
	if err != nil {
		Err("Failed to send GMStreamPacket over connection: %v", err)
		return err
	}

	return writer.Flush()
}
