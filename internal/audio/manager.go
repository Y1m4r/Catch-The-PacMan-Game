package audio

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

// AudioManager handles loading and playing sound effects.
type AudioManager struct {
	sounds        map[string]*beep.Buffer // Store preloaded sound buffers
	format        beep.Format             // Store the format (assuming all WAVs have same format)
	mu            sync.Mutex              // Protect access to sounds map
	isInitialized bool
}

// NewAudioManager creates a new audio manager and initializes the speaker.
func NewAudioManager() (*AudioManager, error) {
	am := &AudioManager{
		sounds: make(map[string]*beep.Buffer),
	}

	// Initialize speaker (needs to be done only once)
	// Choose a sample rate appropriate for your sounds
	// 44100Hz or 48000Hz are common
	sampleRate := beep.SampleRate(44100)
	err := speaker.Init(sampleRate, sampleRate.N(time.Second/10)) // Adjust buffer size if needed
	if err != nil {
		// Log the error but don't necessarily stop the game - maybe run without sound
		log.Printf("Failed to initialize audio speaker: %v. Audio will be disabled.", err)
		return am, nil // Return manager but indicate failure via isInitialized
	}
	am.isInitialized = true
	am.format.SampleRate = sampleRate // Store sample rate
	log.Println("Audio speaker initialized successfully.")

	return am, nil
}

// LoadSound loads a WAV file into a buffer.
func (am *AudioManager) LoadSound(name, filepath string) error {
	if !am.isInitialized {
		return fmt.Errorf("audio manager not initialized, cannot load sound")
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.sounds[name]; exists {
		log.Printf("Sound '%s' already loaded.", name)
		return nil // Avoid reloading
	}

	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("could not open sound file %s: %w", filepath, err)
	}
	// Don't defer close here, streamer needs it open

	streamer, format, err := wav.Decode(f) // Decode closes the file automatically on streamer.Close() or error
	if err != nil {
		return fmt.Errorf("could not decode wav file %s: %w", filepath, err)
	}
	// Note: Using streamer directly might cause issues if played multiple times concurrently.
	// Loading into a buffer allows reusing the sound data safely.

	// Assuming first loaded sound dictates the format for the speaker
	if am.format.NumChannels == 0 {
		am.format = format
		// Re-initialize speaker if format mismatch? Beep handles resampling usually.
		log.Printf("Audio format set based on '%s': SampleRate %d, Channels %d, Precision %d",
			name, format.SampleRate, format.NumChannels, format.Precision)
	} else if am.format != format {
		log.Printf("Warning: Sound '%s' format (%v) differs from expected (%v). Beep will attempt resampling.", name, format, am.format)
		// Beep usually handles resampling, but good to be aware.
	}

	buffer := beep.NewBuffer(am.format) // Create buffer with the initialized format
	buffer.Append(streamer)
	streamer.Close() // Close the streamer after appending to buffer

	am.sounds[name] = buffer
	log.Printf("Loaded sound '%s' from %s", name, filepath)
	return nil
}

// PlaySound plays a preloaded sound by name.
func (am *AudioManager) PlaySound(name string) {
	if !am.isInitialized {
		return // Silently fail if audio isn't working
	}

	am.mu.Lock()
	buffer, ok := am.sounds[name]
	am.mu.Unlock() // Unlock after getting buffer reference

	if !ok {
		log.Printf("Attempted to play unloaded sound: %s", name)
		return
	}

	// Create a streamer from the buffer's data. This allows playing the sound
	// from the beginning each time PlaySound is called, even if it's already playing.
	soundStreamer := buffer.Streamer(0, buffer.Len())

	// Play the sound without blocking. Speaker handles concurrency.
	speaker.Play(soundStreamer)
}

// Close cleans up audio resources (if necessary in future).
func (am *AudioManager) Close() {
	// Speaker doesn't have an explicit Close function in current Beep versions.
	// Resources are usually managed globally or via context.
	log.Println("Audio Manager closed (speaker cleanup is implicit).")
}
