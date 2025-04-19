package persistence

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Y1m4r/Catch-The-PacMan-Game/internal/game" // Adjust path
)

// SaveGame writes the current state of the game to a text file.
func SaveGame(g *game.Game, filepath string) error {
	// Ensure the saves directory exists
	if err := os.MkdirAll("assets/saves", 0755); err != nil {
		return fmt.Errorf("could not create saves directory: %w", err)
	}

	// Use the game's thread-safe method to get data
	level, totalBounces, pacmanData := g.GetDataForSave()

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating save file %s: %w", filepath, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write header: Level and Total Bounces
	_, err = fmt.Fprintf(writer, "%d\n", level)
	if err != nil {
		return fmt.Errorf("error writing level to save file: %w", err)
	}
	_, err = fmt.Fprintf(writer, "%d\n", totalBounces) // Save total bounces too!
	if err != nil {
		return fmt.Errorf("error writing total bounces to save file: %w", err)
	}

	// Write each Pacman's state
	for _, pData := range pacmanData {
		// Format: diameter<tab>posX<tab>posY<tab>waitTimeMs<tab>direction<tab>subDirection<tab>bounces<tab>isStopped
		line := fmt.Sprintf("%.2f\t%.2f\t%.2f\t%d\t%c\t%d\t%d\t%t\n",
			pData.Diameter, // Save diameter
			pData.PosX,
			pData.PosY,
			pData.WaitTimeMs,
			pData.Direction,
			pData.SubDirection, // Save sub-direction
			pData.Bounces,
			pData.IsStopped,
		)
		_, err = writer.WriteString(line)
		if err != nil {
			return fmt.Errorf("error writing pacman data to save file: %w", err)
		}
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("error flushing save file buffer: %w", err)
	}

	log.Printf("Game state saved to %s", filepath)
	return nil
}

// LoadGame reads a game state from a text file.
// Returns a *partial* game object containing loaded state.
func LoadGame(filepath string) (*game.Game, error) {
	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("save file '%s' not found", filepath)
		}
		return nil, fmt.Errorf("error opening save file %s: %w", filepath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	level := -1
	totalBounces := -1
	pacmans := []*game.Pacman{}
	idCounter := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip potential blank lines or comments if any were accidentally saved
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// First non-blank line is the level
		if level == -1 {
			levelVal, err := strconv.Atoi(line)
			if err != nil {
				return nil, fmt.Errorf("line %d: expected level number, got '%s': %w", lineNum, line, err)
			}
			level = levelVal
			continue
		}

		// Second non-blank line is total bounces
		if totalBounces == -1 {
			bouncesVal, err := strconv.Atoi(line)
			if err != nil {
				return nil, fmt.Errorf("line %d: expected total bounces number, got '%s': %w", lineNum, line, err)
			}
			totalBounces = bouncesVal
			continue
		}

		// Subsequent lines are Pac-Man definitions
		parts := strings.Split(line, "\t")
		// Expected format: diameter, posX, posY, waitTimeMs, direction, subDirection, bounces, isStopped (8 fields)
		if len(parts) < 8 {
			log.Printf("Warning line %d: Invalid Pac-Man save data in %s. Expected 8 tab-separated fields, got %d. Skipping line.", lineNum, filepath, len(parts))
			continue
		}

		diameter, errDia := strconv.ParseFloat(parts[0], 64)
		posX, errX := strconv.ParseFloat(parts[1], 64)
		posY, errY := strconv.ParseFloat(parts[2], 64)
		waitTimeMs, errWait := strconv.Atoi(parts[3])
		directionStr := parts[4]
		subDirection, errSubDir := strconv.Atoi(parts[5])
		bounces, errBounce := strconv.Atoi(parts[6])
		isStoppedStr := strings.ToLower(parts[7]) // Case-insensitive boolean

		if errDia != nil || errX != nil || errY != nil || errWait != nil || errSubDir != nil || errBounce != nil {
			log.Printf("Warning line %d: Error parsing values for saved Pac-Man in %s. Skipping line. Errors: %v,%v,%v,%v,%v,%v",
				lineNum, filepath, errDia, errX, errY, errWait, errSubDir, errBounce)
			continue
		}

		var direction rune
		if len(directionStr) > 0 {
			d := strings.ToUpper(directionStr)[0]
			if d == game.DirHorizontal || d == game.DirVertical {
				direction = rune(d)
			} else {
				log.Printf("Warning line %d: Invalid direction '%s' for loaded Pac-Man in %s. Defaulting to Horizontal.", lineNum, directionStr, filepath)
				direction = game.DirHorizontal // Default on load error?
			}
		} else {
			log.Printf("Warning line %d: Missing direction for loaded Pac-Man in %s. Defaulting to Horizontal.", lineNum, filepath)
			direction = game.DirHorizontal
		}

		if subDirection != 1 && subDirection != -1 {
			log.Printf("Warning line %d: Invalid sub-direction '%d' for loaded Pac-Man in %s. Defaulting to 1.", lineNum, subDirection, filepath)
			subDirection = 1
		}

		isStopped := (isStoppedStr == "true" || isStoppedStr == "1")

		radius := diameter / 2.0
		if radius <= 0 {
			log.Printf("Warning line %d: Invalid diameter/radius (<=0) for loaded Pac-Man in %s. Skipping.", lineNum, filepath)
			continue
		}

		pacman := game.NewPacman(idCounter, radius, posX, posY, direction, subDirection, waitTimeMs, bounces, isStopped)
		pacmans = append(pacmans, pacman)
		idCounter++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading save file %s: %w", filepath, err)
	}

	if level == -1 || totalBounces == -1 {
		return nil, fmt.Errorf("save file %s did not contain valid level or bounce data", filepath)
	}

	// Return a *partial* Game struct containing the loaded state
	loadedGame := &game.Game{
		Level:        level,
		TotalBounces: totalBounces,
		Pacmans:      pacmans,
	}

	log.Printf("Loaded game state from %s: Level %d, Bounces %d, %d Pacmans.", filepath, level, totalBounces, len(pacmans))

	return loadedGame, nil
}
