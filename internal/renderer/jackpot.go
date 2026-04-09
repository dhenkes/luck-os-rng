package renderer

import (
	"fmt"
	"time"

	"github.com/dhenkes/luck-os-rng/internal/model"
)

// JackpotFrames returns celebration animation frames based on win tier.
// Frames use Tag "jackpot" so SSE can route them to a dedicated display area.
func JackpotFrames(tier model.JackpotTier) []Frame {
	switch tier {
	case model.JackpotSmall:
		return smallWinFrames()
	case model.JackpotBig:
		return bigWinFrames()
	case model.JackpotMega:
		return megaWinFrames()
	case model.JackpotUltra:
		return ultraWinFrames()
	default:
		return nil
	}
}

func smallWinFrames() []Frame {
	lines := []string{
		"",
		fmt.Sprintf("    %s%s* * * W I N ! * * *%s", Bold, Green, Reset),
		"",
	}
	return []Frame{{
		Content: "\n" + joinLines(lines),
		Lines:   lines,
		Delay:   800 * time.Millisecond,
		Tag:     "jackpot",
	}}
}

func bigWinFrames() []Frame {
	art := []string{
		"",
		fmt.Sprintf("  %s%s __        __ ___  _   _  _  %s", Bold, Yellow, Reset),
		fmt.Sprintf("  %s%s \\ \\      / /|_ _|| \\ | || | %s", Bold, Yellow, Reset),
		fmt.Sprintf("  %s%s  \\ \\ /\\ / /  | | |  \\| || | %s", Bold, Yellow, Reset),
		fmt.Sprintf("  %s%s   \\ V  V /   | | | |\\  ||_| %s", Bold, Yellow, Reset),
		fmt.Sprintf("  %s%s    \\_/\\_/   |___||_| \\_|(_) %s", Bold, Yellow, Reset),
		"",
	}
	return []Frame{{
		Content: "\n" + joinLines(art),
		Lines:   art,
		Delay:   1200 * time.Millisecond,
		Tag:     "jackpot",
	}}
}

func megaWinFrames() []Frame {
	patterns := [3][]string{
		{
			"",
			fmt.Sprintf("  %s%s       *                %s", Bold, Yellow, Reset),
			fmt.Sprintf("  %s%s      ***               %s", Bold, Yellow, Reset),
			fmt.Sprintf("  %s%s     *****              %s", Bold, Yellow, Reset),
			fmt.Sprintf("  %s%s    *JACKPOT*           %s", Bold, Red, Reset),
			fmt.Sprintf("  %s%s     *****              %s", Bold, Yellow, Reset),
			fmt.Sprintf("  %s%s      ***               %s", Bold, Yellow, Reset),
			fmt.Sprintf("  %s%s       *                %s", Bold, Yellow, Reset),
			"",
		},
		{
			"",
			fmt.Sprintf("  %s%s     .    *              %s", Bold, Cyan, Reset),
			fmt.Sprintf("  %s%s    * * .               %s", Bold, Cyan, Reset),
			fmt.Sprintf("  %s%s   . * * *              %s", Bold, Cyan, Reset),
			fmt.Sprintf("  %s%s    *JACKPOT*           %s", Bold, Red, Reset),
			fmt.Sprintf("  %s%s   . * * *              %s", Bold, Cyan, Reset),
			fmt.Sprintf("  %s%s    * * .               %s", Bold, Cyan, Reset),
			fmt.Sprintf("  %s%s     .    *              %s", Bold, Cyan, Reset),
			"",
		},
		{
			"",
			fmt.Sprintf("  %s%s    * . *               %s", Bold, Yellow, Reset),
			fmt.Sprintf("  %s%s   ..***.               %s", Bold, Yellow, Reset),
			fmt.Sprintf("  %s%s  * ***** *             %s", Bold, Yellow, Reset),
			fmt.Sprintf("  %s%s    *JACKPOT*           %s", Bold, Red, Reset),
			fmt.Sprintf("  %s%s  * ***** *             %s", Bold, Yellow, Reset),
			fmt.Sprintf("  %s%s   ..***.               %s", Bold, Yellow, Reset),
			fmt.Sprintf("  %s%s    * . *               %s", Bold, Yellow, Reset),
			"",
		},
	}
	frames := make([]Frame, 3)
	for i, p := range patterns {
		var content string
		if i == 0 {
			content = "\n" + joinLines(p)
		} else {
			content = redraw(p, len(patterns[i-1]))
		}
		frames[i] = Frame{
			Content: content,
			Lines:   p,
			Delay:   400 * time.Millisecond,
			Tag:     "jackpot",
		}
	}
	return frames
}

func ultraWinFrames() []Frame {
	var frames []Frame
	colors := []string{Yellow, Red, Magenta, Cyan, Green}
	var prevLines int
	for i := 0; i < 5; i++ {
		c := colors[i%len(colors)]
		art := []string{
			"",
			fmt.Sprintf("  %s%s *  *  *  *  *  *  *  *  *  *  *  * %s", Bold, c, Reset),
			fmt.Sprintf("  %s%s*                                    *%s", Bold, c, Reset),
			fmt.Sprintf("  %s%s   M  E  G  A   J  A  C  K  P  O  T %s", Bold, c, Reset),
			fmt.Sprintf("  %s%s*                                    *%s", Bold, c, Reset),
			fmt.Sprintf("  %s%s *  *  *  *  *  *  *  *  *  *  *  * %s", Bold, c, Reset),
			"",
		}
		var content string
		if i == 0 {
			content = "\n" + joinLines(art)
		} else {
			content = redraw(art, prevLines)
		}
		prevLines = len(art)
		frames = append(frames, Frame{
			Content: content,
			Lines:   art,
			Delay:   300 * time.Millisecond,
			Tag:     "jackpot",
		})
	}
	return frames
}

func joinLines(lines []string) string {
	result := ""
	for i, l := range lines {
		if i > 0 {
			result += "\n"
		}
		result += l
	}
	return result
}
