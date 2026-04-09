package renderer

import (
	"testing"

	"github.com/dhenkes/luck-os-rng/internal/model"
)

func TestJackpotFramesNone(t *testing.T) {
	frames := JackpotFrames(model.JackpotNone)
	if frames != nil {
		t.Fatal("JackpotNone should produce nil frames")
	}
}

func TestJackpotFramesSmall(t *testing.T) {
	frames := JackpotFrames(model.JackpotSmall)
	if len(frames) != 1 {
		t.Fatalf("small win: %d frames, want 1", len(frames))
	}
	if len(frames[0].Lines) == 0 {
		t.Fatal("small win frame has no lines")
	}
}

func TestJackpotFramesBig(t *testing.T) {
	frames := JackpotFrames(model.JackpotBig)
	if len(frames) != 1 {
		t.Fatalf("big win: %d frames, want 1", len(frames))
	}
}

func TestJackpotFramesMega(t *testing.T) {
	frames := JackpotFrames(model.JackpotMega)
	if len(frames) != 3 {
		t.Fatalf("mega win: %d frames, want 3", len(frames))
	}
	for i, f := range frames {
		if len(f.Lines) == 0 {
			t.Fatalf("mega frame %d has no lines", i)
		}
		if f.Delay == 0 {
			t.Fatalf("mega frame %d should have delay", i)
		}
	}
}

func TestJackpotFramesUltra(t *testing.T) {
	frames := JackpotFrames(model.JackpotUltra)
	if len(frames) != 5 {
		t.Fatalf("ultra win: %d frames, want 5", len(frames))
	}
}
