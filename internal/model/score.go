package model

import "sort"

const MaxHighScores = 10

// Score holds the player's name and their score (number of bounces).
// Needs to be exported for gob encoding/decoding.
type Score struct {
	Name  string
	Score int // Lower is better (fewer bounces)
}

// ByScore implements sort.Interface for []Score based on the Score field (ascending).
type ByScore []Score

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].Score < a[j].Score }

// AddScore adds a new score to the list, keeps it sorted, and trims to MaxHighScores.
// Returns the updated list and true if the score was added (i.e., it made the top list).
// Now operates on []model.Score.
func AddScore(scores []Score, newScore Score) ([]Score, bool) {
	// Check if the new score is better than the worst score currently in the top 10
	// or if the list isn't full yet.
	shouldAdd := false
	if len(scores) < MaxHighScores {
		shouldAdd = true
	} else {
		// Sort scores temporarily to check against the worst if needed
		// (Only sort if the list is full to avoid unnecessary sorting)
		tempScores := make([]Score, len(scores))
		copy(tempScores, scores)
		sort.Sort(ByScore(tempScores))
		if newScore.Score < tempScores[len(tempScores)-1].Score {
			shouldAdd = true
		}
	}

	if shouldAdd {
		scores = append(scores, newScore)
		sort.Sort(ByScore(scores)) // Sort by score ascending

		// Keep only the top MaxHighScores
		if len(scores) > MaxHighScores {
			scores = scores[:MaxHighScores]
		}

		// Check if the added score is actually still in the list after trimming
		for _, s := range scores {
			if s == newScore { // Compare value since it's a simple struct
				return scores, true
			}
		}
		// If we reach here, the score was added but immediately trimmed
		return scores, false
	}

	return scores, false // Score wasn't good enough
}
