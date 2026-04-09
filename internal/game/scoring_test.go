package game

import (
	"testing"

	"github.com/dhenkes/luck-os-rng/internal/model"
)

func TestBetMultiplier(t *testing.T) {
	tests := []struct {
		bet  model.BetTier
		want int
	}{
		{model.BetLow, 1},
		{model.BetMedium, 3},
		{model.BetHigh, 10},
		{model.BetMax, 100},
	}
	for _, tt := range tests {
		got := BetMultiplier(tt.bet)
		if got != tt.want {
			t.Fatalf("BetMultiplier(%v) = %d, want %d", tt.bet, got, tt.want)
		}
	}
}

func TestScoreSlotsWin(t *testing.T) {
	result := model.SlotsResult{
		Mode:       model.SlotsStandard,
		Paylines:   []model.Payline{{LineIndex: 0, Symbol: "CH", Count: 3}},
		Multiplier: 2,
	}
	points, isWin := ScoreSlots(result, model.BetLow)
	if !isWin {
		t.Fatal("expected win")
	}
	// 1 payline * 100 * 2 multiplier * 1 bet = 200
	if points != 200 {
		t.Fatalf("points = %d, want 200", points)
	}
}

func TestScoreSlotsLoss(t *testing.T) {
	result := model.SlotsResult{
		Mode:       model.SlotsStandard,
		Multiplier: 1,
	}
	points, isWin := ScoreSlots(result, model.BetLow)
	if isWin {
		t.Fatal("expected loss")
	}
	if points != 0 {
		t.Fatalf("points = %d, want 0", points)
	}
}

func TestScoreSlotsWithBet(t *testing.T) {
	result := model.SlotsResult{
		Mode:       model.SlotsStandard,
		Paylines:   []model.Payline{{LineIndex: 0, Symbol: "CH", Count: 3}},
		Multiplier: 1,
	}
	points, _ := ScoreSlots(result, model.BetMax)
	// 1 * 100 * 1 * 100 = 10000
	if points != 10000 {
		t.Fatalf("points = %d, want 10000", points)
	}
}

func TestScoreSlotsNonStandard(t *testing.T) {
	result := model.SlotsResult{Mode: model.SlotsMinMax}
	points, isWin := ScoreSlots(result, model.BetLow)
	if !isWin {
		t.Fatal("non-standard modes always win")
	}
	if points != 50 {
		t.Fatalf("points = %d, want 50", points)
	}
}

func TestUpdateEngagement(t *testing.T) {
	old := model.EngagementState{Streak: 2, Score: 500, History: "WW"}
	newEng := UpdateEngagement(old, true, 300)
	if newEng.Streak != 3 {
		t.Fatalf("Streak = %d, want 3", newEng.Streak)
	}
	if newEng.Score != 800 {
		t.Fatalf("Score = %d, want 800", newEng.Score)
	}
	if newEng.History != "WWW" {
		t.Fatalf("History = %q, want WWW", newEng.History)
	}
}

func TestUpdateEngagementLoss(t *testing.T) {
	old := model.EngagementState{Streak: 5, Score: 1000, History: "WWWWW"}
	newEng := UpdateEngagement(old, false, 0)
	if newEng.Streak != 0 {
		t.Fatalf("Streak = %d, want 0", newEng.Streak)
	}
	if newEng.Score != 1000 {
		t.Fatalf("Score = %d, want 1000 (unchanged)", newEng.Score)
	}
	if newEng.History != "WWWWWL" {
		t.Fatalf("History = %q, want WWWWWL", newEng.History)
	}
}

func TestUpdateEngagementHistoryTruncation(t *testing.T) {
	old := model.EngagementState{History: "WWWWWWWWWWWWWWWW"} // 16 chars
	newEng := UpdateEngagement(old, true, 100)
	if len(newEng.History) != 16 {
		t.Fatalf("History length = %d, want 16", len(newEng.History))
	}
	if newEng.History[15] != 'W' {
		t.Fatal("last char should be W")
	}
}

func TestDetectStreak(t *testing.T) {
	tests := []struct {
		history  string
		wantKind string
		wantN    int
	}{
		{"", "", 0},
		{"W", "", 0},
		{"WW", "HOT", 2},
		{"LWWWW", "HOT", 4},
		{"LL", "COLD", 2},
		{"WLLL", "COLD", 3},
		{"WL", "", 0},
	}
	for _, tt := range tests {
		kind, n := DetectStreak(tt.history)
		if kind != tt.wantKind || n != tt.wantN {
			t.Fatalf("DetectStreak(%q) = (%q, %d), want (%q, %d)", tt.history, kind, n, tt.wantKind, tt.wantN)
		}
	}
}

func TestDetermineJackpotTier(t *testing.T) {
	// No win.
	noWin := model.SlotsResult{Mode: model.SlotsStandard, Multiplier: 1}
	if tier := DetermineJackpotTier(noWin, model.BetLow); tier != model.JackpotNone {
		t.Fatalf("no win: tier = %v, want None", tier)
	}

	// Small win.
	small := model.SlotsResult{Mode: model.SlotsStandard, Multiplier: 1,
		Paylines: []model.Payline{{LineIndex: 0, Symbol: "CH", Count: 3}}}
	if tier := DetermineJackpotTier(small, model.BetLow); tier != model.JackpotSmall {
		t.Fatalf("small win: tier = %v, want Small", tier)
	}

	// Big win (multiplier >= 3).
	big := model.SlotsResult{Mode: model.SlotsStandard, Multiplier: 3,
		Paylines: []model.Payline{{LineIndex: 0, Symbol: "CH", Count: 3}}}
	if tier := DetermineJackpotTier(big, model.BetLow); tier != model.JackpotBig {
		t.Fatalf("big win: tier = %v, want Big", tier)
	}

	// Mega win (multiplier >= 5).
	mega := model.SlotsResult{Mode: model.SlotsStandard, Multiplier: 5,
		Paylines: []model.Payline{{LineIndex: 0, Symbol: "CH", Count: 3}}}
	if tier := DetermineJackpotTier(mega, model.BetLow); tier != model.JackpotMega {
		t.Fatalf("mega win: tier = %v, want Mega", tier)
	}

	// Ultra win (multiplier >= 7).
	ultra := model.SlotsResult{Mode: model.SlotsStandard, Multiplier: 7,
		Paylines: []model.Payline{{LineIndex: 0, Symbol: "CH", Count: 3}}}
	if tier := DetermineJackpotTier(ultra, model.BetLow); tier != model.JackpotUltra {
		t.Fatalf("ultra win: tier = %v, want Ultra", tier)
	}
}

