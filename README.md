# Gossip

**End-to-End Encrypted Hub & Spoke Communications Network with Mesh P2P WebRTC Voice and Video Calling**

[![Go Version](https://img.shields.io/badge/Go-1.22.1+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-0.1.0-orange.svg)](gossip-client/main.go)

## Overview

Gossip is a secure, decentralized communications platform that combines the reliability of a hub-and-spoke architecture with the privacy and efficiency of peer-to-peer (P2P) connections. Built with Go and modern web technologies, Gossip provides end-to-end encrypted messaging, voice, and video calling capabilities.

### Key Features

- 🔐 **End-to-End Encryption**: All communications are encrypted using OpenPGP with elliptic curve cryptography
- 🛡️ **Zero-Knowledge Architecture**: Server cannot decrypt client communications despite facilitating key exchange
- 🌐 **Hub & Spoke Architecture**: Centralized signaling server with decentralized P2P connections
- 📞 **WebRTC Voice/Video**: High-quality real-time communication using WebRTC with unlimited participant group calling*
- 🎯 **Mesh P2P Network**: Direct peer-to-peer connections for optimal performance
- 💬 **Ephemeral Messaging**: Messages with configurable expiration times
- 🎨 **Modern UI**: Beautiful, responsive interface built with Svelte and Tailwind CSS
- 🔧 **Cross-Platform**: Desktop application built with Wails framework
- 🚀 **Real-time**: Instant messaging and live audio/video streaming

## Architecture

### System Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Gossip Client │    │  Gossip Server  │    │  Gossip Common  │
│   (Desktop App) │◄──►│  (Signaling)    │◄──►│  (Shared Lib)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   WebRTC P2P    │    │   TCP Server    │    │   Crypto Utils  │
│   Connections   │    │   (Port 1720)   │    │   & Packets     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Communication Flow

1. **Signaling Phase**: Clients connect to the central server for initial handshake
2. **Zero-Knowledge Key Exchange**: Public keys are exchanged through the server using client-side encryption
3. **P2P Establishment**: WebRTC connections are established directly between peers
4. **Encrypted Communication**: All data flows through encrypted P2P channels

## Project Structure

```
gossip/
├── gossip-client/          # Desktop client application
│   ├── frontend/          # Svelte-based UI
│   │   ├── src/
│   │   │   ├── App.svelte     # Main application component
│   │   │   ├── Chat.svelte    # Chat interface
│   │   │   └── components/    # UI components
│   │   └── package.json
│   ├── app.go             # Main application logic
│   ├── stream.go          # WebRTC streaming implementation
│   ├── audio.go           # Audio capture and playback
│   ├── events.go          # Event handling
│   └── main.go            # Application entry point
├── gossip-server/         # Central signaling server
│   ├── main.go            # Server entry point
│   └── events.go          # Server event handling
└── gossip-common/         # Shared utilities and protocols
    ├── crypto.go          # Encryption/decryption functions
    ├── packet.go          # Network packet definitions
    └── utility.go         # Common utilities
```

## Technology Stack

### Backend
- **Go 1.22.1+**: Core application logic
- **Wails v2**: Desktop application framework
- **WebRTC v4**: Real-time communication
- **OpenPGP**: End-to-end encryption
- **Pion**: WebRTC implementation

### Frontend
- **Svelte 4**: Reactive UI framework
- **Tailwind CSS**: Utility-first styling
- **Skeleton UI**: Component library
- **Vite**: Build tool and dev server

### Networking
- **TCP**: Server-client signaling
- **WebRTC**: Peer-to-peer connections
- **STUN/TURN**: NAT traversal

## Installation

### Prerequisites

- Go 1.22.1 or higher
- Node.js 18+ and npm
- Git

### Building from Source

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/gossip.git
   cd gossip
   ```

2. **Build the common library**
   ```bash
   cd gossip-common
   go mod download
   go build
   cd ..
   ```

3. **Build the server**
   ```bash
   cd gossip-server
   go mod download
   go build -o gossip-server
   cd ..
   ```

4. **Build the client**
   ```bash
   cd gossip-client
   go mod download
   cd frontend
   npm install
   npm run build
   cd ..
   wails build
   ```

## Usage

### Starting the Server

```bash
cd gossip-server
./gossip-server [options]

Options:
  -h string     IP to listen on (default "127.0.0.1")
  -p int        Port to listen on (default 1720)
  -k string     Server password (default "anonymous")
  -d            Enable debug logging
  -l            Enable connection logging
```

### Running the Client

1. **Launch the application**
   ```bash
   cd gossip-client
   ./gossip-client
   ```

2. **Connect to a server**
   - Enter server host (default: 127.0.0.1)
   - Enter port (default: 1720)
   - Enter server password
   - Enter your username

3. **Start communicating**
   - Send messages in channels
   - Initiate voice/video calls
   - Manage your settings

## Security Features

### Zero-Knowledge Key Exchange

Gossip implements a sophisticated zero-knowledge key exchange mechanism that ensures the server cannot decrypt any client-to-client communications, even though it facilitates the key distribution process.

#### Key Exchange Process

1. **Client Registration**: When a client connects, it sends its public key to the server
2. **Server Encryption**: The server encrypts each client's public key with the recipient's public key before forwarding
3. **Client Decryption**: Each client decrypts received public keys using their private key
4. **Secure Communication**: Clients can now communicate directly using each other's public keys

#### Zero-Knowledge Properties

- **Server Blindness**: The server never sees unencrypted public keys
- **Client-Side Encryption**: All key encryption/decryption happens on client devices
- **No Key Storage**: The server only stores encrypted keys temporarily
- **Forward Secrecy**: Even if server is compromised, past communications remain secure

#### Implementation Details

```go
// Server encrypts client keys before forwarding
encryptedKey, err := gossip_common.GWEncrypt(clientPublicKey, recipientPublicKey)
keyPacket := gossip_common.NewSignalPacketFromData("ckp", recipientID, senderID, encryptedKey)

// Client decrypts received keys
decryptedKey, err := gossip_common.GWDecrypt(packet.Payload)
publicKeys[packet.Sender] = decryptedKey
```

### Encryption

- **OpenPGP Implementation**: Uses elliptic curve cryptography
- **Key Generation**: Automatic key pair generation on first run
- **Message Encryption**: All messages encrypted with recipient's public key
- **Perfect Forward Secrecy**: WebRTC provides additional security layer
- **Zero-Knowledge Key Exchange**: Server cannot decrypt client-to-client communications

### Authentication

- **Server Authentication**: Password-based server access
- **Client Identification**: Unique client IDs for message routing
- **Channel Security**: Encrypted channel communications

### Privacy

- **No Message Storage**: Messages are ephemeral and not stored
- **P2P Communication**: Direct connections bypass server for data
- **Configurable Expiration**: Messages auto-delete after set time
- **Zero-Knowledge Architecture**: Server acts as encrypted relay without decryption capability

## Configuration

### Server Configuration

The server creates configuration files in the system temp directory:

- `gossip_server.name`: Server display name
- `gossip_channels.list`: Available channels (one per line)

### Client Configuration

Client settings are stored in `gossip_settings.json`:

```json
{
  "selectedTheme": "wintry",
  "defaultUsername": "your_username",
  "defaultHost": "127.0.0.1",
  "defaultPort": "1720"
}
```

## Development

### Project Setup

1. **Install Wails CLI**
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

2. **Install frontend dependencies**
   ```bash
   cd gossip-client/frontend
   npm install
   ```

3. **Run in development mode**
   ```bash
   cd gossip-client
   wails dev
   ```

### Code Structure

#### Packet Types

- **GMSigPacket**: Signaling messages for WebRTC negotiation
- **GMDataPacket**: Encrypted data messages with metadata
- **GMStreamPacket**: Real-time audio/video streaming data

#### Security Packet Flow

```
Client A                    Server                    Client B
   |                          |                          |
   |-- GMSigPacket(grtng) -->|                          |
   |<-- GMSigPacket(hru) ----|                          |
   |-- GMSigPacket(ighru) -->|                          |
   |<-- GMSigPacket(ig) -----|                          |
   |-- GMSigPacket(gmk) -----|                          |
   |<-- GMSigPacket(ckp) ----|-- GMSigPacket(ckp) ----->|
   |                          |                          |
   |-- GMDataPacket(msg) ---->|-- GMDataPacket(msg) ---->|
   |                          |                          |
```

**Packet Security Levels:**
- **Type 0 (GMSigPacket)**: Server can read metadata, encrypted payload
- **Type 1 (GMDataPacket)**: Server cannot decrypt, only forwards
- **Type 2 (GMStreamPacket)**: Server cannot decrypt, targeted forwarding

#### Key Functions

- **Crypto**: `GWEncrypt()`, `GWDecrypt()`, `GenerateKeys()`
- **Networking**: `SendSignalPacket()`, `SendDataPacket()`
- **WebRTC**: `SendOfferToClient()`, `HandleOffer()`, `SendAudioToChannels()`

## API Reference

### Client Methods

```go
// Connection management
Boot(host string, port int, username string, password string) error
Disconnect() error

// Messaging
SendMessage(message string, expiry int64, channel string) error

// Voice/Video
StartRecording()
StopRecording()
ToggleGoMute()
ToggleGoDeaf()
UpdateCallID(callerID string)

// Settings
LoadSettings() (Settings, error)
SaveSettings(settings Settings) error
```

### Server Commands

#### Authentication & Key Exchange
- `grtng`: Client greeting with public key
- `hru`: Server response with server public key
- `ighru`: Encrypted password authentication
- `ig`: Server confirmation of authentication
- `gmk`: Request for all client public keys
- `ckp`: Encrypted client public key packet
- `cup`: Encrypted channel update packet
- `eok`: End of keys transmission
- `rmk`: Remove client key notification

#### Messaging & Calls
- `msg`: Encrypted text message
- `gmp`: Get call participants
- `start_call`: Initialize call session
- `offer`: WebRTC offer
- `answer`: WebRTC answer
- `ice`: ICE candidate exchange
- `hang-up`: Terminate call session

#### Security Classification
- **Server-Readable**: `grtng`, `hru`, `gmk`, `eok`, `rmk` (metadata only)
- **Server-Encrypted**: `ighru`, `ig`, `ckp`, `cup` (server can decrypt for routing)
- **Client-Only**: `msg`, `offer`, `answer`, `ice` (server cannot decrypt)

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go coding standards
- Add tests for new functionality
- Update documentation for API changes
- Ensure encryption security is maintained

## Troubleshooting

### Common Issues

1. **Connection Failed**
   - Verify server is running
   - Check firewall settings
   - Confirm host/port configuration

2. **Audio/Video Issues**
   - Check microphone/camera permissions
   - Verify WebRTC support in browser
   - Test with STUN/TURN servers

3. **Encryption Errors**
   - Regenerate keys if corrupted
   - Check OpenPGP implementation
   - Verify key exchange process

### Debug Mode

Enable debug logging for detailed troubleshooting:

```bash
# Server
./gossip-server -d -l

# Client (set debugLogging = true in main.go)
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Wails](https://wails.io/) - Desktop application framework
- [Pion WebRTC](https://github.com/pion/webrtc) - WebRTC implementation
- [Svelte](https://svelte.dev/) - Frontend framework
- [OpenPGP](https://www.openpgp.org/) - Encryption standard

## * Roadmap

- [ ] Video calling support
- [ ] File sharing capabilities
- [ ] Mobile client development
- [ ] Advanced encryption options
- [ ] Server clustering support
- [ ] Plugin system
- [ ] API documentation

---

**Gossip** - Secure communications for the modern world. Built open and free by m4yc3x under the [Open Research Initiative for Web Technologies Foundation](ori.wtf)
