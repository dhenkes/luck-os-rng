package model

import (
	"net/url"
	"testing"
)

func TestParseBetTier(t *testing.T) {
	tests := []struct {
		input string
		want  BetTier
	}{
		{"", BetLow},
		{"low", BetLow},
		{"medium", BetMedium},
		{"high", BetHigh},
		{"max", BetMax},
		{"invalid", BetLow},
	}
	for _, tt := range tests {
		got := ParseBetTier(tt.input)
		if got != tt.want {
			t.Fatalf("ParseBetTier(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestBetTierString(t *testing.T) {
	tests := []struct {
		tier BetTier
		want string
	}{
		{BetLow, "low"},
		{BetMedium, "medium"},
		{BetHigh, "high"},
		{BetMax, "max"},
	}
	for _, tt := range tests {
		if got := tt.tier.String(); got != tt.want {
			t.Fatalf("BetTier(%d).String() = %q, want %q", tt.tier, got, tt.want)
		}
	}
}

func TestBetTierLabel(t *testing.T) {
	if l := BetMax.Label(); l != "MAX (100x)" {
		t.Fatalf("BetMax.Label() = %q", l)
	}
}

func TestParseEngagementState(t *testing.T) {
	q := url.Values{}
	q.Set("s", "3")
	q.Set("sc", "1500")
	q.Set("h", "WWLW")
	q.Set("u", "high")
	q.Set("bet", "medium")

	eng := ParseEngagementState(q)
	if eng.Streak != 3 {
		t.Fatalf("Streak = %d, want 3", eng.Streak)
	}
	if eng.Score != 1500 {
		t.Fatalf("Score = %d, want 1500", eng.Score)
	}
	if eng.History != "WWLW" {
		t.Fatalf("History = %q, want WWLW", eng.History)
	}
	if eng.Unlocked != "high" {
		t.Fatalf("Unlocked = %q, want high", eng.Unlocked)
	}
	if eng.Bet != BetMedium {
		t.Fatalf("Bet = %v, want medium", eng.Bet)
	}
}

func TestParseEngagementStateDefaults(t *testing.T) {
	eng := ParseEngagementState(url.Values{})
	if eng.Streak != 0 || eng.Score != 0 || eng.History != "" || eng.Unlocked != "" || eng.Bet != BetLow {
		t.Fatal("empty query should produce zero-value engagement state")
	}
}

func TestParseEngagementStateHistoryTruncation(t *testing.T) {
	q := url.Values{}
	q.Set("h", "WWWWWWWWWWWWWWWWWWWW") // 20 chars
	eng := ParseEngagementState(q)
	if len(eng.History) != 16 {
		t.Fatalf("History length = %d, want 16", len(eng.History))
	}
}

func TestEngagementQueryString(t *testing.T) {
	eng := EngagementState{Streak: 3, Score: 1500, History: "WWW", Unlocked: "high", Bet: BetMedium}
	qs := eng.QueryString()
	if qs == "" {
		t.Fatal("expected non-empty query string")
	}
	// Should contain all non-default values.
	for _, want := range []string{"s=3", "sc=1500", "h=WWW", "u=high", "bet=medium"} {
		found := false
		for _, part := range splitQS(qs) {
			if part == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("query string %q missing %q", qs, want)
		}
	}
}

func TestEngagementQueryStringDefaults(t *testing.T) {
	eng := EngagementState{}
	if qs := eng.QueryString(); qs != "" {
		t.Fatalf("zero-value engagement should produce empty query string, got %q", qs)
	}
}

func splitQS(qs string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(qs); i++ {
		if qs[i] == '&' {
			parts = append(parts, qs[start:i])
			start = i + 1
		}
	}
	parts = append(parts, qs[start:])
	return parts
}

func TestHasUnlock(t *testing.T) {
	eng := EngagementState{Unlocked: "high,insane"}
	if !eng.HasUnlock("high") {
		t.Fatal("expected HasUnlock(high) = true")
	}
	if !eng.HasUnlock("insane") {
		t.Fatal("expected HasUnlock(insane) = true")
	}
	if eng.HasUnlock("ultra") {
		t.Fatal("expected HasUnlock(ultra) = false")
	}
}

func TestHasUnlockEmpty(t *testing.T) {
	eng := EngagementState{}
	if eng.HasUnlock("high") {
		t.Fatal("empty unlocked should return false")
	}
}

func TestAddUnlock(t *testing.T) {
	eng := EngagementState{}
	eng.AddUnlock("high")
	if eng.Unlocked != "high" {
		t.Fatalf("Unlocked = %q, want high", eng.Unlocked)
	}
	eng.AddUnlock("insane")
	if eng.Unlocked != "high,insane" {
		t.Fatalf("Unlocked = %q, want high,insane", eng.Unlocked)
	}
	// Adding duplicate should not change.
	eng.AddUnlock("high")
	if eng.Unlocked != "high,insane" {
		t.Fatalf("Unlocked = %q after duplicate add", eng.Unlocked)
	}
}
