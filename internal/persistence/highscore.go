package persistence

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	// Use your module path for model
	"github.com/Y1m4r/Catch-The-PacMan-Game/internal/model" // <--- IMPORT model
	// NO LONGER import game here!
)

// SaveHighScores takes []model.Score
func SaveHighScores(scores []model.Score, filepath string) error { // <--- Parameter uses model.Score
	if err := os.MkdirAll("assets/highscores", 0755); err != nil {
		return fmt.Errorf("could not create highscores directory: %w", err)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating high score file %s: %w", filepath, err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	// Encode the []model.Score slice
	err = encoder.Encode(scores) // <--- Encode the slice directly
	if err != nil {
		return fmt.Errorf("error encoding high scores to %s: %w", filepath, err)
	}
	log.Printf("High scores saved successfully to %s (%d entries)", filepath, len(scores))
	return nil
}

// LoadHighScores returns []model.Score
func LoadHighScores(filepath string) ([]model.Score, error) { // <--- Return type uses model.Score
	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("High score file %s not found. Returning empty list.", filepath)
			return []model.Score{}, nil // <--- Return empty model.Score slice
		}
		return nil, fmt.Errorf("error opening high score file %s: %w", filepath, err)
	}
	defer file.Close()

	var scores []model.Score // <--- USE model.Score
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&scores) // <--- Decode into model.Score slice

	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Printf("Reached end of high score file %s (or file was empty).", filepath)
			if scores == nil {
				scores = []model.Score{} // <--- Ensure non-nil model.Score slice
			}
			return scores, nil // <--- Return model.Score slice
		}
		return nil, fmt.Errorf("error decoding high scores from %s: %w", filepath, err)
	}

	log.Printf("High scores loaded successfully from %s (%d entries)", filepath, len(scores))
	return scores, nil // <--- Return model.Score slice
}
