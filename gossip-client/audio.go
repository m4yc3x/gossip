package main

import (
	"fmt"
	"sync"

	"github.com/gen2brain/malgo"
)

// Recorder encapsulates the audio recording logic.
type Recorder struct {
	deviceConfig malgo.DeviceConfig
	device       *malgo.Device
	buffer       [][]byte
	mutex        sync.Mutex
}

// NewRecorder creates a new Recorder with the specified sample rate.
func NewRecorder() *Recorder {
	return &Recorder{
		deviceConfig: malgo.DefaultDeviceConfig(malgo.Capture),
	}
}

// initDevice initializes the malgo device for audio capture.
func (r *Recorder) initDevice() error {
	r.deviceConfig.Capture.Format = malgo.FormatS16
	r.deviceConfig.Capture.Channels = 1
	r.deviceConfig.SampleRate = 44100
	r.deviceConfig.Alsa.NoMMap = 1

	// Create a malgo context
	allocatedCtx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize context: %v", err)
	}
	defer allocatedCtx.Uninit()

	// Define the callback using malgo.DeviceCallbacks
	callbacks := malgo.DeviceCallbacks{
		Data: func(output, input []byte, framecount uint32) {
			r.mutex.Lock()
			defer r.mutex.Unlock()

			// Copy input to a new slice to avoid overwriting during async operations
			chunk := make([]byte, len(input))
			copy(chunk, input)
			//r.buffer = append(r.buffer, chunk)
			if !muted && !deafened {
				SendAudioToChannels(chunk)
			}
		},
	}

	// Initialize the device with the correct parameters using the Context field of AllocatedContext
	r.device, err = malgo.InitDevice(allocatedCtx.Context, r.deviceConfig, callbacks)
	if err != nil {
		return fmt.Errorf("failed to initialize device: %v", err)
	}

	return nil
}

// Start begins the audio recording process.
func (r *Recorder) Start() error {
	if err := r.initDevice(); err != nil {
		return err
	}

	if err := r.device.Start(); err != nil {
		return fmt.Errorf("failed to start device: %v", err)
	}

	return nil
}

// Stop halts the audio recording process.
func (r *Recorder) Stop() error {
	if err := r.device.Stop(); err != nil {
		return fmt.Errorf("failed to stop device: %v", err)
	}
	return nil
}

// GetBuffer returns the recorded audio data.
func (r *Recorder) GetBuffer() [][]byte {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.buffer
}
