package main

import (
	"fmt"
	"sync"

	"gossip_common"

	"github.com/gen2brain/malgo"
)

// Player encapsulates the audio playback logic.
type Player struct {
	deviceConfig malgo.DeviceConfig
	device       *malgo.Device
	buffer       [][]byte
	currentIndex int
	mutex        sync.Mutex
}

// NewPlayer creates a new Player with the specified sample rate.
func NewPlayer() *Player {
	return &Player{
		deviceConfig: malgo.DefaultDeviceConfig(malgo.Playback),
	}
}

// initDevice initializes the malgo device for audio playback.
func (p *Player) initDevice(buffer [][]byte) error {
	p.deviceConfig.Playback.Format = malgo.FormatS16
	p.deviceConfig.Playback.Channels = 1
	p.deviceConfig.SampleRate = 44100
	p.deviceConfig.Alsa.NoMMap = 1
	p.buffer = buffer

	// Create a malgo context
	allocatedCtx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize context: %v", err)
	}
	defer allocatedCtx.Uninit()

	var maxSample int16 = 0 // Running maximum sample value
	var normalizationFactor float64 = 1.0
	const silenceThreshold int16 = 500 // Threshold below which we consider the audio as silent

	// Define the callback using malgo.DeviceCallbacks
	callbacks := malgo.DeviceCallbacks{
		Data: func(output, input []byte, framecount uint32) {
			p.mutex.Lock()
			defer p.mutex.Unlock()

			var totalBytesCopied int = 0

			// Update running maximum sample value
			for _, data := range p.buffer {
				for i := 0; i < len(data); i += 2 {
					sample := int16(data[i]) | int16(data[i+1])<<8
					if abs(sample) > maxSample {
						maxSample = abs(sample)
					}
				}
			}

			// Apply normalization only if the max sample is above the silence threshold
			if maxSample > silenceThreshold {
				normalizationFactor = 32767.0 / float64(maxSample) * 0.78
			} else {
				normalizationFactor = 1.0 // No normalization if below threshold
			}

			// Normalize and copy data
			for totalBytesCopied < len(output) {
				if p.currentIndex >= len(p.buffer) {
					// Fill the rest of the output with silence if no more data
					for i := totalBytesCopied; i < len(output); i++ {
						output[i] = 0
					}
					break
				}

				data := p.buffer[p.currentIndex]
				remainingOutputSpace := len(output) - totalBytesCopied
				bytesToCopy := min(len(data), remainingOutputSpace)

				for i := 0; i < bytesToCopy; i += 2 {
					sample := int16(data[i]) | int16(data[i+1])<<8
					normalizedSample := int16(float64(sample) * normalizationFactor)
					output[totalBytesCopied+i] = byte(normalizedSample & 0xFF)
					output[totalBytesCopied+i+1] = byte((normalizedSample >> 8) & 0xFF)
				}

				totalBytesCopied += bytesToCopy

				if bytesToCopy < len(data) {
					p.buffer[p.currentIndex] = data[bytesToCopy:]
				} else {
					p.currentIndex++
				}
			}
		},
	}

	// Initialize the device with the correct parameters using the Context field of AllocatedContext
	p.device, err = malgo.InitDevice(allocatedCtx.Context, p.deviceConfig, callbacks)
	if err != nil {
		return fmt.Errorf("failed to initialize device: %v", err)
	}

	return nil
}

// Helper function to find the absolute value of an int16
func abs(x int16) int16 {
	if x < 0 {
		return -x
	}
	return x
}

// Helper function to find minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Start begins the audio playback process.
func (p *Player) Start() error {
	if err := p.device.Start(); err != nil {
		return fmt.Errorf("failed to start device: %v", err)
	}
	return nil
}

// Stop halts the audio playback process.
func (p *Player) Stop() error {
	if err := p.device.Stop(); err != nil {
		return fmt.Errorf("failed to stop device: %v", err)
	}
	return nil
}

func (p *Player) AddToBuffer(buffer []byte) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if deafened {
		return
	}

	// Decrypt the incoming buffer before adding it to the player's buffer
	decryptedBuffer, err := gossip_common.GWDecrypt(buffer)
	if err != nil {
		gossip_common.Err("Failed to decrypt buffer: %v", err)
		return
	}

	p.buffer = append(p.buffer, decryptedBuffer)
}
