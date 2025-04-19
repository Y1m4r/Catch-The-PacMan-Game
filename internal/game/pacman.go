package game

import (
	"sync"
	"time"
)

const (
	DirHorizontal = 'H'
	DirVertical   = 'V'
	// Speed pixels per second - adjust as needed
	baseSpeed = 60.0
)

// Pacman represents a single Pac-Man character in the game.
type Pacman struct {
	ID           int
	Radius       float64
	PosX         float64 // Center X
	PosY         float64 // Center Y
	Speed        float64 // Pixels per second
	Direction    rune    // 'H' or 'V'
	SubDirection int     // 1 for right/down, -1 for left/up
	IsStopped    bool
	WaitTimeMs   int // Original config value, might influence speed or animation
	Bounces      int // Bounces against walls or other Pacmans

	// Animation state
	animFrame    int
	lastAnimTime time.Time
	animInterval time.Duration

	// Mutex to protect this Pacman's state during concurrent access
	// This is kept internal to the Pacman methods.
	mu sync.Mutex
}

// NewPacman creates a new Pacman instance from configuration data.
func NewPacman(id int, radius, posX, posY float64, direction rune, subDirection int, waitTimeMs, bounces int, isStopped bool) *Pacman {
	// Example speed calculation: faster if waitTimeMs is lower
	speed := baseSpeed * (100.0 / (float64(waitTimeMs) + 1)) // Avoid division by zero, adjust formula as needed

	return &Pacman{
		ID:           id,
		Radius:       radius,
		PosX:         posX,
		PosY:         posY,
		Speed:        speed,
		Direction:    direction,
		SubDirection: subDirection,
		IsStopped:    isStopped,
		WaitTimeMs:   waitTimeMs,
		Bounces:      bounces,
		animFrame:    0,
		lastAnimTime: time.Now(),
		animInterval: 150 * time.Millisecond, // Adjust animation speed
	}
}

// Update moves the Pacman and handles animation frame switching.
// screenWidth and screenHeight define the play area boundaries.
// dt is the delta time (time since last update) in seconds.
// Returns the number of bounces that occurred during this update.
func (p *Pacman) Update(dt float64, screenWidth, screenHeight float64) (bounces int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.IsStopped {
		return 0
	}

	// --- Animation ---
	if time.Since(p.lastAnimTime) > p.animInterval {
		p.animFrame = (p.animFrame + 1) % 2 // Cycle between 0 and 1
		p.lastAnimTime = time.Now()
	}

	// --- Movement ---
	distance := p.Speed * dt
	bounced := false
	startBounces := p.Bounces

	if p.Direction == DirHorizontal {
		p.PosX += distance * float64(p.SubDirection)
		// Check boundaries
		if p.PosX-p.Radius < 0 && p.SubDirection == -1 {
			p.PosX = p.Radius // Snap to boundary
			p.SubDirection *= -1
			bounced = true
		} else if p.PosX+p.Radius > screenWidth && p.SubDirection == 1 {
			p.PosX = screenWidth - p.Radius // Snap to boundary
			p.SubDirection *= -1
			bounced = true
		}
	} else { // DirVertical
		p.PosY += distance * float64(p.SubDirection)
		// Check boundaries
		if p.PosY-p.Radius < 0 && p.SubDirection == -1 {
			p.PosY = p.Radius // Snap to boundary
			p.SubDirection *= -1
			bounced = true
		} else if p.PosY+p.Radius > screenHeight && p.SubDirection == 1 {
			p.PosY = screenHeight - p.Radius // Snap to boundary
			p.SubDirection *= -1
			bounced = true
		}
	}

	if bounced {
		p.Bounces++
	}

	return p.Bounces - startBounces // Return bounces occurred *in this step*
}

// Bounce changes the Pacman's direction due to collision with another Pacman.
// It increments the bounce count and returns true.
func (p *Pacman) Bounce() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.IsStopped {
		return false // Cannot bounce if stopped
	}
	p.SubDirection *= -1
	p.Bounces++

	// Small positional nudge to prevent immediate re-collision
	nudge := 1.1 // Adjust nudge factor if needed
	if p.Direction == DirHorizontal {
		p.PosX += nudge * float64(p.SubDirection)
	} else {
		p.PosY += nudge * float64(p.SubDirection)
	}

	return true
}

// Stop marks the Pacman as stopped and returns true if it was running.
func (p *Pacman) Stop() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.IsStopped {
		p.IsStopped = true
		return true // Was running, now stopped
	}
	return false // Was already stopped
}

// IsClicked checks if the given coordinates (cx, cy) are inside the Pacman.
// Safe for concurrent read access if needed, but Stop() must be called via Game.
func (p *Pacman) IsClicked(cx, cy float64) bool {
	p.mu.Lock() // Lock needed to read position safely
	defer p.mu.Unlock()
	// Simple circle collision check
	dx := p.PosX - cx
	dy := p.PosY - cy
	distanceSq := dx*dx + dy*dy
	return distanceSq < p.Radius*p.Radius && !p.IsStopped
}

// GetData returns a thread-safe copy of the Pacman's current state for drawing or saving.
func (p *Pacman) GetData() (posX, posY, radius float64, animFrame int, isStopped bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.PosX, p.PosY, p.Radius, p.animFrame, p.IsStopped
}

// GetDataForSave returns a thread-safe copy of the Pacman's state relevant for saving.
func (p *Pacman) GetDataForSave() (radius, posX, posY float64, waitTimeMs, subDirection, bounces int, direction rune, isStopped bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// Diameter is often stored in config, but radius is used internally. Save radius for consistency? Let's save diameter.
	return p.Radius * 2, p.PosX, p.PosY, p.WaitTimeMs, p.SubDirection, p.Bounces, p.Direction, p.IsStopped
}

// CheckCollision detects collision with another Pacman.
// Takes the *other* Pacman's locked data to avoid deadlocks.
func (p *Pacman) CheckCollision(otherPosX, otherPosY, otherRadius float64) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.IsStopped {
		return false
	} // Cannot collide if stopped

	dx := p.PosX - otherPosX
	dy := p.PosY - otherPosY
	distSq := dx*dx + dy*dy
	radiiSum := p.Radius + otherRadius

	return distSq < radiiSum*radiiSum
}

// GetStateForCollisionCheck returns necessary data under lock for collision checking.
func (p *Pacman) GetStateForCollisionCheck() (posX, posY, radius float64, isStopped bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.PosX, p.PosY, p.Radius, p.IsStopped
}
