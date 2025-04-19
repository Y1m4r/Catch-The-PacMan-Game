package game

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Y1m4r/Catch-The-PacMan-Game/internal/audio" // Adjust import path
	"github.com/Y1m4r/Catch-The-PacMan-Game/internal/model" //
)

// GameState represents the possible states of the game screen.
type GameState int

const (
	StateStarting GameState = iota // Initial state, loading maybe
	StatePlaying
	StateGameOver
	StateEnteringHighScore // Waiting for player name input
	StateHallOfFame        // Displaying high scores
)

// Game represents the overall game state and logic controller.
type Game struct {
	Pacmans      []*Pacman
	Level        int
	TotalBounces int
	ScreenWidth  float64
	ScreenHeight float64
	CurrentState GameState

	HighScores      []model.Score // Loaded high scores for the current level
	highScorePath   string        // Path to save/load high scores for this level
	saveGamePath    string        // Path to save the current game state
	levelConfigPath string        // Path of the loaded level

	lastUpdateTime time.Time
	deltaTime      float64 // Time since last frame in seconds

	// Player name input buffer (for high score entry)
	playerNameInput []rune
	isNewHighScore  bool // Flag if the current score qualifies for high scores

	audioManager *audio.AudioManager // Reference to the audio manager

	// Mutex to protect shared game state (Pacmans slice, TotalBounces, CurrentState, HighScores)
	mu sync.RWMutex // Allows multiple readers (Draw) or one writer (Update, HandleClick)

}

func (g *Game) IsNewHighScorePending() (any, any) {
	panic("unimplemented")
}

func (g *Game) ResetToStart() {
	panic("unimplemented")
}

// NewGame initializes a new game state, but doesn't load a level yet.
func NewGame(screenWidth, screenHeight float64, audioMgr *audio.AudioManager) *Game {
	g := &Game{
		Level:        -1, // No level loaded initially
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
		CurrentState: StateStarting,
		Pacmans:      []*Pacman{},
		HighScores:   []model.Score{},
		audioManager: audioMgr,
	}
	return g
}

// RequestLoadLevel triggers the loading of a level configuration.
// It acquires the write lock to modify game state safely.
func (g *Game) RequestLoadLevel(level int, configPath string, loadFunc func(string) (*Game, error)) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	log.Printf("Requesting load level %d from %s", level, configPath)
	loadedGameData, err := loadFunc(configPath)
	if err != nil {
		log.Printf("Error loading level config %s: %v", configPath, err)
		return fmt.Errorf("failed to load level config '%s': %w", configPath, err)
	}

	// Transfer loaded data to the current game instance
	g.Level = loadedGameData.Level
	g.Pacmans = loadedGameData.Pacmans
	g.TotalBounces = loadedGameData.TotalBounces // Usually 0 for new level, but loader might set it
	g.CurrentState = StatePlaying
	g.levelConfigPath = configPath
	g.highScorePath = fmt.Sprintf("assets/highscores/highscores_%d.gob", g.Level)
	g.saveGamePath = fmt.Sprintf("assets/saves/savegame_%d.txt", g.Level) // Or a generic quicksave path
	g.playerNameInput = []rune{}
	g.isNewHighScore = false

	// Call the injected loader function (which now returns []model.Score)
	if loadHighScoresFunc != nil {
		loadedScores, err := loadHighScoresFunc(g.highScorePath)
		if err != nil {
			log.Printf("Could not load high scores for level %d (%s): %v. Starting fresh.", g.Level, g.highScorePath, err)
			g.HighScores = []model.Score{} // <--- USE model.Score
		} else {
			g.HighScores = loadedScores // <--- Assign loaded []model.Score
			log.Printf("Loaded %d high scores for level %d", len(g.HighScores), g.Level)
		}
	} else {
		log.Printf("Warning: High score loading function not set.")
		g.HighScores = []model.Score{} // <--- USE model.Score
	}

	g.lastUpdateTime = time.Now()
	log.Printf("Level %d loaded successfully. Starting game.", g.Level)
	if g.audioManager != nil {
		// g.audioManager.PlaySound("level_start")
	}

	return nil
}

// RequestLoadSavedGame triggers loading from a save file.
func (g *Game) RequestLoadSavedGame(savePath string, loadFunc func(string) (*Game, error)) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	log.Printf("Requesting load saved game from %s", savePath)
	loadedGameData, err := loadFunc(savePath)
	if err != nil {
		log.Printf("Error loading saved game %s: %v", savePath, err)
		return fmt.Errorf("failed to load saved game '%s': %w", savePath, err)
	}

	// Transfer loaded data
	g.Level = loadedGameData.Level
	g.Pacmans = loadedGameData.Pacmans
	g.TotalBounces = loadedGameData.TotalBounces
	g.CurrentState = StatePlaying
	// Determine paths based on loaded level
	g.levelConfigPath = fmt.Sprintf("assets/levels/level_%d.txt", g.Level) // Assume standard naming
	g.highScorePath = fmt.Sprintf("assets/highscores/highscores_%d.gob", g.Level)
	g.saveGamePath = savePath // Keep the path we loaded from
	g.playerNameInput = []rune{}
	g.isNewHighScore = false

	// Call the injected loader function (which now returns []model.Score)
	if loadHighScoresFunc != nil {
		loadedScores, err := loadHighScoresFunc(g.highScorePath)
		if err != nil {
			log.Printf("Could not load high scores for loaded level %d (%s): %v. Starting fresh.", g.Level, g.highScorePath, err)
			g.HighScores = []model.Score{} // <--- USE model.Score
		} else {
			g.HighScores = loadedScores // <--- Assign loaded []model.Score
		}
	} else {
		log.Printf("Warning: High score loading function not set.")
		g.HighScores = []model.Score{} // <--- USE model.Score
	}

	g.lastUpdateTime = time.Now()
	log.Printf("Saved game loaded successfully. Resuming level %d.", g.Level)
	return nil
}

// RequestSaveGame triggers saving the current game state.
func (g *Game) RequestSaveGame(saveFunc func(*Game, string) error) error {
	g.mu.RLock() // Use Read Lock initially to check state
	if g.CurrentState != StatePlaying || g.Level < 0 {
		g.mu.RUnlock()
		log.Println("Cannot save game: Not currently playing a level.")
		return fmt.Errorf("cannot save game: not playing")
	}
	currentSavePath := g.saveGamePath // Get path while read-locked
	g.mu.RUnlock()                    // Release read lock before calling save function

	log.Printf("Requesting save game to %s", currentSavePath)
	// The saveFunc will need to acquire necessary locks (Read lock on Game, locks on Pacmans)
	// Pass 'g' itself so saveFunc can access data via public methods or direct fields (if within same package)
	err := saveFunc(g, currentSavePath)
	if err != nil {
		log.Printf("Error saving game state to %s: %v", currentSavePath, err)
		return fmt.Errorf("failed to save game: %w", err)
	}

	log.Printf("Game state saved successfully to %s", currentSavePath)
	return nil
}

// Update proceeds the game state by one step.
// It handles Pacman movement, collisions, state transitions, and input for name entry.
func (g *Game) Update() {
	g.mu.Lock() // Lock for writing state
	defer g.mu.Unlock()

	now := time.Now()
	g.deltaTime = now.Sub(g.lastUpdateTime).Seconds()
	g.lastUpdateTime = now

	// Only update game elements if playing
	if g.CurrentState != StatePlaying {
		return // Don't update Pacmans, bounces etc. if not playing
	}

	if g.Level < 0 {
		log.Println("Warning: Game Update called but no level loaded.")
		return // Should not happen if state transitions are correct
	}

	allStopped := true
	bouncesThisFrame := 0

	// --- Pacman Movement & Edge Bouncing ---
	for _, p := range g.Pacmans {
		bounces := p.Update(g.deltaTime, g.ScreenWidth, g.ScreenHeight) // Update handles its own lock
		bouncesThisFrame += bounces
		_, _, _, _, stopped := p.GetData() // Safely get stopped status
		if !stopped {
			allStopped = false
		}
	}

	// --- Pacman-to-Pacman Collision ---
	numPacmans := len(g.Pacmans)
	for i := 0; i < numPacmans; i++ {
		p1 := g.Pacmans[i]
		p1PosX, p1PosY, p1Radius, p1Stopped := p1.GetStateForCollisionCheck()
		if p1Stopped {
			continue
		}

		for j := i + 1; j < numPacmans; j++ {
			p2 := g.Pacmans[j]
			p2PosX, p2PosY, p2Radius, p2Stopped := p2.GetStateForCollisionCheck()
			if p2Stopped {
				continue
			}

			// Check collision using the retrieved safe data
			dx := p1PosX - p2PosX
			dy := p1PosY - p2PosY
			distSq := dx*dx + dy*dy
			radiiSum := p1Radius + p2Radius

			if distSq > 0 && distSq < radiiSum*radiiSum { // distSq > 0 avoids collision with self if logic flawed
				// Collision detected! Bounce both Pacmans.
				// The Bounce method handles internal state update & bounce count.
				bounced1 := p1.Bounce()
				bounced2 := p2.Bounce()
				if bounced1 {
					bouncesThisFrame++
				}
				if bounced2 {
					bouncesThisFrame++
				}
				if bounced1 || bounced2 {
					// Play bounce sound maybe? Limit frequency?
					if g.audioManager != nil {
						// g.audioManager.PlaySound("pacman_bounce") // Add a bounce sound
					}
				}
			}
		}
	}

	g.TotalBounces += bouncesThisFrame

	// Check for game over condition
	if allStopped {
		g.CurrentState = StateGameOver
		log.Printf("Game Over! Final Bounces: %d", g.TotalBounces)
		if g.audioManager != nil {
			// g.audioManager.PlaySound("level_up") // Or a specific game over sound
		}
		// Check if score qualifies for Hall of Fame
		_, g.isNewHighScore = model.AddScore(g.HighScores, model.Score{Score: g.TotalBounces}) // Check without adding yet
		if g.isNewHighScore {
			log.Println("New High Score achieved!")
			g.CurrentState = StateEnteringHighScore // Transition to name entry state
			g.playerNameInput = []rune{}            // Clear input buffer
		}
	}
}

// HandleClick checks if any Pacman was clicked at (x, y) and stops it.
// Acquires necessary locks.
func (g *Game) HandleClick(x, y float64) {
	g.mu.Lock() // Need write lock to potentially modify Pacman state
	defer g.mu.Unlock()

	if g.CurrentState != StatePlaying {
		return // Ignore clicks if not playing
	}

	for _, p := range g.Pacmans {
		// IsClicked is safe, checks bounds and if already stopped
		if p.IsClicked(x, y) {
			wasRunning := p.Stop() // Stop method handles its own mutex and state change
			if wasRunning && g.audioManager != nil {
				g.audioManager.PlaySound("pacman_death") // Play sound on successful stop
			}
			break // Assume only one Pacman can be clicked at a time
		}
	}
}

// HandleTextInput processes character input during the high score entry state.
func (g *Game) HandleTextInput(chars []rune) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.CurrentState != StateEnteringHighScore {
		return
	}
	// Append new characters, limit name length if desired
	if len(g.playerNameInput) < 15 { // Limit name length
		g.playerNameInput = append(g.playerNameInput, chars...)
	}
}

// HandleBackspace removes the last character during high score entry.
func (g *Game) HandleBackspace() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.CurrentState != StateEnteringHighScore && len(g.playerNameInput) > 0 {
		g.playerNameInput = g.playerNameInput[:len(g.playerNameInput)-1]
	}
}

// HandleEnter confirms the entered name and saves the high score.
func (g *Game) HandleEnter(saveFunc func([]model.Score, string) error) {
	g.mu.Lock() // Acquire write lock
	defer g.mu.Unlock()

	if g.CurrentState != StateEnteringHighScore {
		return
	}

	playerName := string(g.playerNameInput)
	if playerName == "" {
		playerName = "Anonymous" // Default name
	}

	log.Printf("Adding high score: %s - %d", playerName, g.TotalBounces)

	var added bool
	g.HighScores, added = model.AddScore(g.HighScores, model.Score{Name: playerName, Score: g.TotalBounces})

	if added {
		log.Println("Score added to Hall of Fame. Saving...")
		err := saveFunc(g.HighScores, g.highScorePath) // Call the persistence function
		if err != nil {
			log.Printf("Failed to save high scores: %v", err)
			// Maybe inform the user in the UI?
		} else {
			log.Println("High scores saved successfully.")
		}
	} else {
		log.Println("Score was not added (likely pushed out by better scores).")
	}

	g.CurrentState = StateHallOfFame // Transition to showing the hall of fame
	g.playerNameInput = []rune{}     // Clear input
}

// --- Data Accessor Methods (Thread-Safe) ---

// GetPacmanData provides data needed for drawing all Pacmans.
func (g *Game) GetPacmanData() []struct {
	PosX, PosY, Radius float64
	AnimFrame          int
	IsStopped          bool
} {
	g.mu.RLock() // Read lock is sufficient
	defer g.mu.RUnlock()

	data := make([]struct {
		PosX, PosY, Radius float64
		AnimFrame          int
		IsStopped          bool
	}, len(g.Pacmans))

	for i, p := range g.Pacmans {
		data[i].PosX, data[i].PosY, data[i].Radius, data[i].AnimFrame, data[i].IsStopped = p.GetData()
	}
	return data
}

// GetGameState provides the current game state and score.
func (g *Game) GetGameState() (state GameState, bounces int, level int) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.CurrentState, g.TotalBounces, g.Level
}

// GetHighScoreData provides data for displaying the Hall of Fame.
func (g *Game) GetHighScoreData() (state GameState, scores []model.Score, currentPlayerName string) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	// Return a copy to prevent modification outside the lock
	scoresCopy := make([]model.Score, len(g.HighScores))
	copy(scoresCopy, g.HighScores)
	return g.CurrentState, scoresCopy, string(g.playerNameInput)
}

// Need to define these somewhere accessible, perhaps passed into NewGame or globally (less ideal)
var loadHighScoresFunc func(filepath string) ([]model.Score, error) = nil // Placeholder
//var saveHighScoresFunc func(scores []Score, filepath string) error = nil // Placeholder - passed into HandleEnter

// SetPersistenceFunctions allows injecting the actual persistence functions
// This avoids import cycles if persistence needs game types.
func SetPersistenceFunctions(loader func(string) ([]model.Score, error)) { // saver func( []Score, string) error) {
	loadHighScoresFunc = loader
	// saveHighScoresFunc = saver // Pass saver to HandleEnter
}

// GetDataForSave provides necessary game state for saving.
func (g *Game) GetDataForSave() (level int, totalBounces int, pacmans []PacmanSaveData) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	level = g.Level
	totalBounces = g.TotalBounces
	pacmans = make([]PacmanSaveData, len(g.Pacmans))
	for i, p := range g.Pacmans {
		// Call the Pacman's safe data retrieval method
		diameter, posX, posY, waitTimeMs, subDirection, bounces, direction, isStopped := p.GetDataForSave()
		pacmans[i] = PacmanSaveData{
			Diameter:     diameter, // Store diameter as per original format
			PosX:         posX,
			PosY:         posY,
			WaitTimeMs:   waitTimeMs,
			Direction:    direction,
			SubDirection: subDirection,
			Bounces:      bounces,
			IsStopped:    isStopped,
		}
	}
	return level, totalBounces, pacmans
}

// PacmanSaveData is a helper struct to hold data for saving a single Pacman.
type PacmanSaveData struct {
	Diameter     float64
	PosX         float64
	PosY         float64
	WaitTimeMs   int
	Direction    rune
	SubDirection int // Added this, seems necessary to restore state
	Bounces      int
	IsStopped    bool
}
