package main

import (
	"log"
	"os"

	"github.com/Y1m4r/Catch-The-PacMan-Game/internal/graphics" // Adjust import path
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Ensure necessary directories exist before game starts
	ensureDir("assets/saves")
	ensureDir("assets/highscores")

	// Create the main game object
	gameInstance, err := graphics.NewEbitenGame()
	if err != nil {
		log.Fatalf("Failed to initialize game: %v", err)
	}

	// Setup Ebiten window
	ebiten.SetWindowSize(graphics.ScreenWidth, graphics.ScreenHeight)
	ebiten.SetWindowTitle("Catch The Pac-Man (Go Version)")
	ebiten.SetWindowClosingHandled(true) // Handle Q key or close button manually if needed

	log.Println("Starting Ebiten game loop...")
	// Run the game loop
	if err := ebiten.RunGame(gameInstance); err != nil {
		// Check if it's the specific "user requested quit" error or something else
		if err.Error() == "user requested quit" {
			log.Println("Game exited normally by user request (Q key).")
		} else {
			log.Printf("Ebiten loop exited with error: %v", err)
		}
	}

	// Clean up resources (like audio speaker) if necessary
	if err := gameInstance.Close(); err != nil {
		log.Printf("Error during game cleanup: %v", err)
	}
	log.Println("Game finished.")
}

// ensureDir creates a directory if it doesn't exist.
func ensureDir(dirName string) {
	err := os.MkdirAll(dirName, 0755) // Use MkdirAll for convenience (creates parents if needed)
	if err != nil {
		// Log the error but don't necessarily make it fatal,
		// as persistence functions might handle the error later.
		log.Printf("Warning: Could not create directory %s: %v", dirName, err)
	}
}
