package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Y1m4r/Catch-The-PacMan-Game/internal/game" // Adjust path
)

// LoadLevelConfig reads a level configuration file and creates a new Game object.
// Note: This returns a *partial* game object containing level data.
// The main game logic should integrate this data into the active game state.
func LoadLevelConfig(filepath string) (*game.Game, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening level file %s: %w", filepath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	level := -1
	pacmans := []*game.Pacman{}
	idCounter := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip blank lines and comments
		}

		// First valid line is the level
		if level == -1 {
			levelVal, err := strconv.Atoi(line)
			if err != nil {
				return nil, fmt.Errorf("line %d: expected level number, got '%s': %w", lineNum, line, err)
			}
			if levelVal < 0 || levelVal > 2 {
				log.Printf("Warning line %d: Invalid level number %d in %s. Defaulting to 0.", lineNum, levelVal, filepath)
				level = 0 // Default or handle as error?
			} else {
				level = levelVal
			}
			continue
		}

		// Subsequent valid lines are Pac-Man definitions
		parts := strings.Split(line, "\t")
		// Expected format: diameter, posX, posY, waitTimeMs, direction, bounces, isStopped (7 fields)
		if len(parts) < 7 { // Be flexible if fields are added later, but require minimum
			log.Printf("Warning line %d: Invalid Pac-Man definition in %s. Expected 7 tab-separated fields, got %d. Skipping line.", lineNum, filepath, len(parts))
			continue
		}

		diameter, errDia := strconv.ParseFloat(parts[0], 64)
		posX, errX := strconv.ParseFloat(parts[1], 64)
		posY, errY := strconv.ParseFloat(parts[2], 64)
		waitTimeMs, errWait := strconv.Atoi(parts[3])
		directionStr := parts[4]
		bounces, errBounce := strconv.Atoi(parts[5])
		isStoppedStr := strings.ToLower(parts[6]) // Case-insensitive boolean

		if errDia != nil || errX != nil || errY != nil || errWait != nil || errBounce != nil {
			log.Printf("Warning line %d: Error parsing numeric values for Pac-Man in %s. Skipping line. Errors: %v,%v,%v,%v,%v",
				lineNum, filepath, errDia, errX, errY, errWait, errBounce)
			continue
		}

		var direction rune
		if len(directionStr) > 0 {
			d := strings.ToUpper(directionStr)[0]
			if d == game.DirHorizontal || d == game.DirVertical {
				direction = rune(d)
			} else {
				log.Printf("Warning line %d: Invalid direction '%s' for Pac-Man in %s. Defaulting to Horizontal.", lineNum, directionStr, filepath)
				direction = game.DirHorizontal
			}
		} else {
			log.Printf("Warning line %d: Missing direction for Pac-Man in %s. Defaulting to Horizontal.", lineNum, filepath)
			direction = game.DirHorizontal
		}

		// Initial sub-direction (Assume 1 for right/down unless specified otherwise - format doesn't include it)
		initialSubDirection := 1

		isStopped := (isStoppedStr == "true" || isStoppedStr == "1")

		radius := diameter / 2.0
		if radius <= 0 {
			log.Printf("Warning line %d: Invalid diameter/radius (<=0) for Pac-Man in %s. Skipping.", lineNum, filepath)
			continue
		}

		pacman := game.NewPacman(idCounter, radius, posX, posY, direction, initialSubDirection, waitTimeMs, bounces, isStopped)
		pacmans = append(pacmans, pacman)
		idCounter++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading level file %s: %w", filepath, err)
	}

	if level == -1 {
		return nil, fmt.Errorf("level file %s did not contain a valid level number", filepath)
	}

	// Return a *partial* Game struct containing the loaded level data
	loadedGame := &game.Game{
		Level:   level,
		Pacmans: pacmans,
		// TotalBounces will be initialized by the main Game logic when loading
	}

	log.Printf("Loaded level %d config from %s with %d Pacmans.", level, filepath, len(pacmans))

	return loadedGame, nil
}
