package game

import "github.com/dhenkes/luck-os-rng/internal/model"

// BetMultiplier returns the score multiplier for a bet tier.
func BetMultiplier(bet model.BetTier) int {
	switch bet {
	case model.BetMedium:
		return 3
	case model.BetHigh:
		return 10
	case model.BetMax:
		return 100
	default:
		return 1
	}
}

// ScoreSlots computes points and win status from a slots result.
func ScoreSlots(result model.SlotsResult, bet model.BetTier) (int, bool) {
	if result.Mode != model.SlotsStandard {
		return 50 * BetMultiplier(bet), true
	}
	if len(result.Paylines) == 0 {
		return 0, false
	}
	points := len(result.Paylines) * 100 * result.Multiplier * BetMultiplier(bet)
	if result.BonusRound {
		points *= 2
	}
	return points, true
}

// UpdateEngagement returns a new engagement state after a game result.
func UpdateEngagement(old model.EngagementState, isWin bool, points int) model.EngagementState {
	eng := old
	if isWin {
		eng.Streak = old.Streak + 1
		eng.Score = old.Score + points
		eng.History = AppendHistory(old.History, 'W')
	} else {
		eng.Streak = 0
		eng.History = AppendHistory(old.History, 'L')
	}
	return eng
}

func AppendHistory(h string, ch byte) string {
	h += string(ch)
	if len(h) > 16 {
		h = h[len(h)-16:]
	}
	return h
}

// DetectStreak scans history from right for consecutive W or L.
// Returns "HOT" / "COLD" and count, or "" and 0.
func DetectStreak(history string) (string, int) {
	if len(history) == 0 {
		return "", 0
	}
	last := history[len(history)-1]
	count := 0
	for i := len(history) - 1; i >= 0; i-- {
		if history[i] == last {
			count++
		} else {
			break
		}
	}
	if count < 2 {
		return "", 0
	}
	if last == 'W' {
		return "HOT", count
	}
	return "COLD", count
}

// DetermineJackpotTier returns the celebration tier for a slots result.
func DetermineJackpotTier(result model.SlotsResult, bet model.BetTier) model.JackpotTier {
	if result.Mode != model.SlotsStandard {
		return model.JackpotNone
	}
	if len(result.Paylines) == 0 {
		return model.JackpotNone
	}

	// Check for ultra-rare outcomes.
	if result.Multiplier >= 7 || (bet == model.BetMax && len(result.Paylines) > 0) {
		return model.JackpotUltra
	}
	if result.Multiplier >= 5 || len(result.Paylines) >= 7 {
		return model.JackpotMega
	}
	if result.Multiplier >= 3 || len(result.Paylines) >= 4 {
		return model.JackpotBig
	}
	return model.JackpotSmall
}

