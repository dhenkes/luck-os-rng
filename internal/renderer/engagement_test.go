package renderer

import (
	"strings"
	"testing"

	"github.com/dhenkes/luck-os-rng/internal/model"
)

func TestEngagementFrameWin(t *testing.T) {
	info := EngagementInfo{
		State: model.EngagementState{
			Streak:  3,
			Score:   1500,
			History: "WWLWWW",
			Bet:     model.BetMedium,
		},
		IsWin:      true,
		HasWinLoss: true,
		Points:     300,
		NextURL:    "localhost:8080/slots?s=3&sc=1500",
		DoubleURL:  "localhost:8080/double?stake=300",
	}
	frame := EngagementFrame(info, 0)

	if len(frame.Lines) == 0 {
		t.Fatal("engagement frame has no lines")
	}

	content := strings.Join(frame.Lines, "\n")

	if !strings.Contains(content, "Score: 1500") {
		t.Fatal("missing score display")
	}
	if !strings.Contains(content, "Streak: W x3") {
		t.Fatal("missing streak display")
	}
	if !strings.Contains(content, "+300 points") {
		t.Fatal("missing points display")
	}
	if !strings.Contains(content, "GO AGAIN") {
		t.Fatal("missing go again URL")
	}
	if !strings.Contains(content, "DOUBLE OR NOTHING") {
		t.Fatal("missing double-or-nothing URL")
	}

	// Meta should contain URLs for browser.
	if frame.Meta["nextURL"] == "" {
		t.Fatal("missing nextURL in Meta")
	}
	if frame.Meta["doubleURL"] == "" {
		t.Fatal("missing doubleURL in Meta")
	}
}

func TestEngagementFrameLoss(t *testing.T) {
	info := EngagementInfo{
		State: model.EngagementState{
			Streak:  0,
			Score:   500,
			History: "WL",
		},
		IsWin:      false,
		HasWinLoss: true,
		Points:     0,
		NextURL:    "localhost:8080/slots",
	}
	frame := EngagementFrame(info, 0)

	content := strings.Join(frame.Lines, "\n")
	if !strings.Contains(content, "Streak: ---") {
		t.Fatal("loss should show streak reset")
	}
	if strings.Contains(content, "DOUBLE OR NOTHING") {
		t.Fatal("loss should not show double-or-nothing")
	}
}

func TestEngagementFrameNoWinLoss(t *testing.T) {
	// Non-slots games: no streak, no double-or-nothing.
	info := EngagementInfo{
		State: model.EngagementState{
			Score: 200,
		},
		IsWin:      true,
		HasWinLoss: false,
		Points:     100,
		NextURL:    "localhost:8080/coinflip",
	}
	frame := EngagementFrame(info, 0)

	content := strings.Join(frame.Lines, "\n")
	if !strings.Contains(content, "Score: 200") {
		t.Fatal("missing score")
	}
	if strings.Contains(content, "Streak") {
		t.Fatal("non-win/loss games should not show streak")
	}
	if strings.Contains(content, "DOUBLE") {
		t.Fatal("non-win/loss games should not show double-or-nothing")
	}
	if !strings.Contains(content, "+100 points") {
		t.Fatal("should still show points earned")
	}
}

func TestEngagementFrameNearMiss(t *testing.T) {
	info := EngagementInfo{
		State:      model.EngagementState{},
		IsWin:      false,
		HasWinLoss: true,
		NearMiss:   []string{"SO CLOSE! One 7s away from TRIPLE SEVENS"},
		NextURL:    "localhost:8080/slots",
	}
	frame := EngagementFrame(info, 0)

	content := strings.Join(frame.Lines, "\n")
	if !strings.Contains(content, "SO CLOSE") {
		t.Fatal("missing near-miss message")
	}
}

func TestHistoryRendering(t *testing.T) {
	info := EngagementInfo{
		State: model.EngagementState{
			History: "WWWWW",
			Score:   100,
			Streak:  5,
		},
		IsWin:      true,
		HasWinLoss: true,
		Points:     100,
		NextURL:    "localhost:8080/slots",
		DoubleURL:  "localhost:8080/double?stake=100",
	}
	frame := EngagementFrame(info, 0)

	content := strings.Join(frame.Lines, "\n")
	if !strings.Contains(content, "Recent:") {
		t.Fatal("missing history display")
	}
	if !strings.Contains(content, "HOT STREAK") {
		t.Fatal("missing hot streak indicator")
	}
}

func TestDoubleOrNothingFrameWin(t *testing.T) {
	frame := DoubleOrNothingFrame(true, 100, 200, "cashout-url", "double-url", nil, 0)
	content := strings.Join(frame.Lines, "\n")
	if !strings.Contains(content, "YOU WON") {
		t.Fatal("missing win message")
	}
	if !strings.Contains(content, "CASH OUT") {
		t.Fatal("missing cash out option")
	}
	if !strings.Contains(content, "DOUBLE AGAIN") {
		t.Fatal("missing double again option")
	}
	if frame.Meta["cashOutURL"] != "cashout-url" {
		t.Fatal("missing cashOutURL in Meta")
	}
}

func TestDoubleOrNothingFrameLoss(t *testing.T) {
	urls := map[string]string{"Slots": "play-again-url"}
	frame := DoubleOrNothingFrame(false, 100, 0, "double-again-url", "", urls, 0)
	content := strings.Join(frame.Lines, "\n")
	if !strings.Contains(content, "YOU LOST") {
		t.Fatal("missing loss message")
	}
	if !strings.Contains(content, "TRY AGAIN") {
		t.Fatal("missing try again option")
	}
	if !strings.Contains(content, "OTHER GAMES") {
		t.Fatal("missing other games option")
	}
	if frame.Meta["nextURL"] != "double-again-url" {
		t.Fatal("missing nextURL in Meta")
	}
}
