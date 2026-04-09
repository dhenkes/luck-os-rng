package model

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type BetTier int

const (
	BetLow BetTier = iota
	BetMedium
	BetHigh
	BetMax
)

func (b BetTier) String() string {
	switch b {
	case BetMedium:
		return "medium"
	case BetHigh:
		return "high"
	case BetMax:
		return "max"
	default:
		return "low"
	}
}

func (b BetTier) Label() string {
	switch b {
	case BetMedium:
		return "MEDIUM (3x)"
	case BetHigh:
		return "HIGH (10x)"
	case BetMax:
		return "MAX (100x)"
	default:
		return "LOW (1x)"
	}
}

func ParseBetTier(s string) BetTier {
	switch s {
	case "medium":
		return BetMedium
	case "high":
		return BetHigh
	case "max":
		return BetMax
	default:
		return BetLow
	}
}

type JackpotTier int

const (
	JackpotNone JackpotTier = iota
	JackpotSmall
	JackpotBig
	JackpotMega
	JackpotUltra
)

// EngagementState is carried across plays in URL query params.
// No database, no cookies -- the URL is the database.
type EngagementState struct {
	Streak   int
	Score    int
	History  string // W/L characters, max 16
	Unlocked string // comma-separated unlock keys
	Bet      BetTier
}

func ParseEngagementState(q url.Values) EngagementState {
	s, _ := strconv.Atoi(q.Get("s"))
	sc, _ := strconv.Atoi(q.Get("sc"))
	h := q.Get("h")
	if len(h) > 16 {
		h = h[len(h)-16:]
	}
	return EngagementState{
		Streak:   s,
		Score:    sc,
		History:  h,
		Unlocked: q.Get("u"),
		Bet:      ParseBetTier(q.Get("bet")),
	}
}

// QueryString returns compact query params for embedding in next-play URLs.
// Only non-default values are included.
func (e EngagementState) QueryString() string {
	var parts []string
	if e.Streak > 0 {
		parts = append(parts, fmt.Sprintf("s=%d", e.Streak))
	}
	if e.Score > 0 {
		parts = append(parts, fmt.Sprintf("sc=%d", e.Score))
	}
	if e.History != "" {
		parts = append(parts, "h="+e.History)
	}
	if e.Unlocked != "" {
		parts = append(parts, "u="+e.Unlocked)
	}
	if e.Bet != BetLow {
		parts = append(parts, "bet="+e.Bet.String())
	}
	return strings.Join(parts, "&")
}

func (e EngagementState) HasUnlock(mode string) bool {
	if e.Unlocked == "" {
		return false
	}
	for _, u := range strings.Split(e.Unlocked, ",") {
		if u == mode {
			return true
		}
	}
	return false
}

func (e *EngagementState) AddUnlock(mode string) {
	if e.HasUnlock(mode) {
		return
	}
	if e.Unlocked == "" {
		e.Unlocked = mode
	} else {
		e.Unlocked += "," + mode
	}
}
