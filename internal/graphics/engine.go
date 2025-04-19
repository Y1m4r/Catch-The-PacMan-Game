package graphics

import (
	"fmt"
	"image/color" // Import color
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil" // For DebugPrint
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	// Use your actual module path
	"github.com/Y1m4r/Catch-The-PacMan-Game/internal/config"
	"github.com/Y1m4r/Catch-The-PacMan-Game/internal/game"
	"github.com/Y1m4r/Catch-The-PacMan-Game/internal/persistence"
)

const (
	ScreenWidth  = 640
	ScreenHeight = 480
)

// Define colors used
var (
	colorBlack    = color.RGBA{0, 0, 0, 255}
	colorWhite    = color.RGBA{255, 255, 255, 255}
	colorYellow   = color.RGBA{R: 255, G: 255, B: 0, A: 255} // Define Yellow
	colorRed      = color.RGBA{R: 255, G: 50, B: 50, A: 255}
	colorGray     = color.Gray{Y: 150}
	colorDarkBlue = color.RGBA{0, 0, 10, 255}
)

// EbitenGame implements ebiten.Game interface and manages the game loop.
type EbitenGame struct {
	GameLogic *game.Game
	Assets    *Assets
}

// NewEbitenGame creates the main game controller for Ebiten.
func NewEbitenGame() (*EbitenGame, error) {
	assets, err := LoadAssets()
	if err != nil {
		return nil, fmt.Errorf("failed to load assets: %w", err)
	}

	coreGame := game.NewGame(float64(ScreenWidth), float64(ScreenHeight), assets.AudioManager)

	// Inject persistence function - Use the correct LoadHighScores from persistence
	game.SetPersistenceFunctions(persistence.LoadHighScores)

	eg := &EbitenGame{
		GameLogic: coreGame,
		Assets:    assets,
	}

	// Initial state is Starting, let Update handle transition based on input
	// No need to explicitly load level 0 here if StateStarting handles it

	return eg, nil
}

// Update proceeds the game state.
func (eg *EbitenGame) Update() error {
	// Use the game's method to get state safely
	state, _, currentLevel := eg.GameLogic.GetGameState()

	// --- Global Input Handling ---
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		return fmt.Errorf("user requested quit")
	}

	// --- Input based on Game State ---
	switch state {
	case game.StatePlaying: // **Use game. prefix**
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			eg.GameLogic.HandleClick(float64(x), float64(y))
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyS) {
			// Pass the actual SaveGame function from persistence
			err := eg.GameLogic.RequestSaveGame(persistence.SaveGame)
			if err != nil {
				log.Printf("Save failed: %v", err)
			} else {
				log.Println("Game Saved (press L to load)")
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyL) {
			if currentLevel >= 0 {
				savePath := fmt.Sprintf("assets/saves/savegame_%d.txt", currentLevel)
				// Pass the actual LoadGame function from persistence
				err := eg.GameLogic.RequestLoadSavedGame(savePath, persistence.LoadGame)
				if err != nil {
					log.Printf("Load failed: %v", err)
				} else {
					log.Println("Game Loaded.")
				}
			} else {
				log.Println("Cannot load: No level currently active to determine save file.")
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
			eg.loadLevel(0)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
			eg.loadLevel(1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
			eg.loadLevel(2)
		}

		eg.GameLogic.Update()

	case game.StateGameOver: // **Use game. prefix**
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if currentLevel >= 0 {
				eg.loadLevel(currentLevel)
			} else {
				eg.loadLevel(0) // Default fallback
			}
		}

	case game.StateEnteringHighScore: // **Use game. prefix**
		// **Use ebiten.InputChars() instead of AppendInputChars**
		inputChars := ebiten.InputChars()
		if len(inputChars) > 0 {
			eg.GameLogic.HandleTextInput(inputChars)
		}
		if repeatingKeyPressed(ebiten.KeyBackspace) { // Allow holding backspace
			eg.GameLogic.HandleBackspace()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			// **Pass the actual SaveHighScores function from persistence**
			eg.GameLogic.HandleEnter(persistence.SaveHighScores)
		}

	case game.StateHallOfFame: // **Use game. prefix**
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			eg.loadLevel(0) // Restart level 0 after viewing scores
		}

	case game.StateStarting: // **Use game. prefix**
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			err := eg.loadLevel(0) // Load level 0 on Enter/Click
			if err != nil {
				log.Printf("Failed to load level 0 on start: %v", err)
				// Optionally, stay in Starting state or show an error
			}
		}
	}

	return nil
}

// Draw renders the game screen based on the current state.
func (eg *EbitenGame) Draw(screen *ebiten.Image) { // **screen is the *ebiten.Image parameter**
	screen.Fill(colorDarkBlue) // Use defined color

	// **Use game's method to get state safely**
	state, bounces, level := eg.GameLogic.GetGameState()

	switch state {
	case game.StateStarting: // **Use game. prefix**
		// **Pass screen to drawText and use defined colors**
		drawText(screen, "Catch The Pac-Man!", ScreenWidth/2, ScreenHeight/3, colorWhite, true)
		drawText(screen, "Press ENTER or Click to Start Level 0", ScreenWidth/2, ScreenHeight/2, colorYellow, true)
		drawText(screen, "Q=Quit", 10, ScreenHeight-20, colorGray, false)

	case game.StatePlaying, game.StateGameOver: // **Use game. prefix**
		pacmanData := eg.GameLogic.GetPacmanData()
		for _, pData := range pacmanData {
			if !pData.IsStopped {
				op := &ebiten.DrawImageOptions{}
				img := eg.Assets.PacmanFrames[pData.AnimFrame]
				bounds := img.Bounds()
				w, h := float64(bounds.Dx()), float64(bounds.Dy())
				op.GeoM.Translate(-w/2, -h/2)
				op.GeoM.Translate(pData.PosX, pData.PosY)
				screen.DrawImage(img, op) // **Draw onto screen**
			}
		}

		// **Pass screen to drawText and use defined colors**
		drawText(screen, fmt.Sprintf("Level: %d", level), 10, 20, colorWhite, false)
		drawText(screen, fmt.Sprintf("Bounces: %d", bounces), ScreenWidth-150, 20, colorWhite, false)
		drawText(screen, "Click PacMan!", ScreenWidth/2, 20, colorYellow, true)
		drawText(screen, "S=Save L=Load Q=Quit F1/F2/F3=Level", 10, ScreenHeight-20, colorGray, false)

		if state == game.StateGameOver { // **Use game. prefix**
			drawText(screen, "GAME OVER!", ScreenWidth/2, ScreenHeight/2-30, colorRed, true)
			drawText(screen, "Press ENTER or Click to Restart", ScreenWidth/2, ScreenHeight/2+10, colorWhite, true)
		}

	case game.StateEnteringHighScore: // **Use game. prefix**
		drawText(screen, fmt.Sprintf("Level: %d", level), 10, 20, colorWhite, false)
		drawText(screen, fmt.Sprintf("Bounces: %d", bounces), ScreenWidth-150, 20, colorWhite, false)

		drawText(screen, "New High Score!", ScreenWidth/2, ScreenHeight/2-60, colorYellow, true)
		drawText(screen, "Enter Your Name:", ScreenWidth/2, ScreenHeight/2-20, colorWhite, true)

		// **Use game's method GetHighScoreData safely**
		_, _, nameInput := eg.GameLogic.GetHighScoreData()
		drawText(screen, nameInput+"_", ScreenWidth/2, ScreenHeight/2+20, colorWhite, true) // Add underscore cursor

		drawText(screen, "Press ENTER to Confirm", ScreenWidth/2, ScreenHeight/2+60, colorWhite, true)

	case game.StateHallOfFame: // **Use game. prefix**
		drawText(screen, "Hall of Fame - Level "+strconv.Itoa(level), ScreenWidth/2, 50, colorYellow, true)

		// **Use game's method GetHighScoreData safely**
		_, scores, _ := eg.GameLogic.GetHighScoreData()
		yPos := 100.0
		for i, score := range scores {
			rankStr := fmt.Sprintf("%d.", i+1)
			scoreStr := fmt.Sprintf("%s  -  %d Bounces", score.Name, score.Score)
			drawText(screen, rankStr, ScreenWidth/3, yPos, colorWhite, false)
			drawText(screen, scoreStr, ScreenWidth/2+20, yPos, colorWhite, false) // Adjust X slightly for alignment
			yPos += 30
		}

		if len(scores) == 0 {
			drawText(screen, "No scores yet!", ScreenWidth/2, ScreenHeight/2, colorGray, true)
		}

		drawText(screen, "Press ENTER or Click to Continue", ScreenWidth/2, ScreenHeight-50, colorWhite, true)
	}
}

// Layout defines the logical screen size.
func (eg *EbitenGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// Helper function to load a specific level
func (eg *EbitenGame) loadLevel(level int) error {
	levelPath := fmt.Sprintf("assets/levels/level_%d.txt", level)
	// Pass the actual LoadLevelConfig function from config
	return eg.GameLogic.RequestLoadLevel(level, levelPath, config.LoadLevelConfig)
}

// Helper function for drawing text
// **Added screen parameter**
func drawText(screen *ebiten.Image, str string, x, y float64, clr color.Color, center bool) {
	// Using DebugPrint for simplicity. Replace with text.Draw for fonts later.
	drawX := x
	if center {
		textWidth := float64(len(str) * 6) // Approximate width for DebugPrint font
		drawX = x - textWidth/2
	}
	// **Use ebitenutil.DebugPrintAt correctly**
	ebitenutil.DebugPrintAt(screen, str, int(drawX), int(y))
}

// repeatingKeyPressed simulates key repeats for keys like backspace.
// From Ebiten examples.
func repeatingKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 30 // Ticks before repeat starts
		interval = 5  // Ticks between repeats
	)
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true // Pressed just now
	}
	if d >= delay && (d-delay)%interval == 0 {
		return true // Repeating
	}
	return false
}

// Close is called when the game is about to exit.
func (eg *EbitenGame) Close() error {
	if eg.Assets != nil && eg.Assets.AudioManager != nil {
		eg.Assets.AudioManager.Close()
	}
	log.Println("EbitenGame closed.")
	return nil
}
