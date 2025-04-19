package graphics

import (
	"fmt"
	"image"
	_ "image/png" // Import for PNG decoding side effects
	"log"
	"os"

	"github.com/Y1m4r/Catch-The-PacMan-Game/internal/audio" // Adjust path
	"github.com/hajimehoshi/ebiten/v2"
)

// Assets holds the loaded graphical and audio resources.
type Assets struct {
	PacmanFrames []*ebiten.Image
	AudioManager *audio.AudioManager
	// Add fonts later if needed
	// Font font.Face
}

// LoadAssets loads all required resources.
func LoadAssets() (*Assets, error) {
	assets := &Assets{
		PacmanFrames: make([]*ebiten.Image, 2), // 2 frames for mouth animation
	}

	// --- Load Images ---
	var err error
	assets.PacmanFrames[0], err = loadImage("assets/images/pacman-0.png")
	if err != nil {
		return nil, fmt.Errorf("failed to load pacman-0.png: %w", err)
	}
	assets.PacmanFrames[1], err = loadImage("assets/images/pacman-1.png")
	if err != nil {
		return nil, fmt.Errorf("failed to load pacman-1.png: %w", err)
	}
	log.Println("Loaded Pac-Man images.")

	// --- Initialize and Load Audio ---
	assets.AudioManager, err = audio.NewAudioManager()
	if err != nil {
		// Non-fatal error, audio manager handles internal state
		log.Printf("Audio Manager initialization partially failed: %v", err)
		// Continue without audio or with limited audio functionality
	}

	// Load sounds even if init failed - LoadSound checks initialization status
	err = assets.AudioManager.LoadSound("pacman_death", "assets/audio/pacman_death.wav")
	if err != nil {
		log.Printf("Warning: failed to load pacman_death sound: %v", err)
	}
	err = assets.AudioManager.LoadSound("level_up", "assets/audio/level_up.wav") // Example: use for game over
	if err != nil {
		log.Printf("Warning: failed to load level_up sound: %v", err)
	}
	// Add other sounds: title_game, pacman_move (if desired)
	// err = assets.AudioManager.LoadSound("title_game", "assets/audio/title_game.wav")
	// if err != nil { log.Printf("Warning: failed to load title_game sound: %v", err) }

	log.Println("Assets loaded successfully.")
	return assets, nil
}

// loadImage is a helper function to load an ebiten.Image from a file path.
func loadImage(path string) (*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}
