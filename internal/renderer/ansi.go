// Package renderer transforms game results into output formats.
package renderer

import (
	"fmt"
	"strings"
	"time"

	"github.com/dhenkes/luck-os-rng/internal/game"
	"github.com/dhenkes/luck-os-rng/internal/model"
)

// Frame is a single snapshot of the terminal display at a point in time.
type Frame struct {
	Content string        // ANSI-encoded content for terminals.
	Lines   []string      // Raw display lines (with ANSI color but no cursor movement).
	Delay   time.Duration
}

// ANSI escape codes.
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Magenta = "\033[35m"
)

// redraw returns the escape sequence to move the cursor back to the start
// of a block with the given number of lines, clear everything, then write
// the new content. On the first frame (prev == 0) it just writes the content.
func redraw(lines []string, prevLineCount int) string {
	var b strings.Builder

	if prevLineCount > 1 {
		// Cursor is on the last line of the previous block.
		// Move up (prevLineCount - 1) lines to reach the first line.
		// \033[{N}F = move cursor to beginning of line N lines up.
		fmt.Fprintf(&b, "\033[%dF", prevLineCount-1)
	} else if prevLineCount == 1 {
		// Previous block was one line; just return to start of current line.
		b.WriteString("\r")
	}
	// prevLineCount == 0: first frame, write from current cursor position.

	for i, line := range lines {
		if i > 0 {
			b.WriteString("\n")
		}
		// \033[2K = clear entire line (regardless of cursor position).
		b.WriteString("\033[2K")
		b.WriteString(line)
	}

	return b.String()
}

// lineCount returns the number of terminal lines a block occupies.
func lineCount(lines []string) int {
	return len(lines)
}

// frameDelay returns the delay for a given frame index with deceleration.
func frameDelay(i int) time.Duration {
	switch {
	case i < 10:
		return 50 * time.Millisecond
	case i < 15:
		return 100 * time.Millisecond
	case i < 20:
		return 200 * time.Millisecond
	case i < 25:
		return 400 * time.Millisecond
	default:
		return 800 * time.Millisecond
	}
}

func colorForRoulette(c model.RouletteColor) string {
	switch c {
	case model.RouletteRed:
		return Red
	case model.RouletteBlack:
		return White
	case model.RouletteGreen:
		return Green
	default:
		return White
	}
}

// RouletteFrames generates ANSI animation frames for a roulette spin.
func RouletteFrames(result model.RouletteResult, cfg model.RouletteConfig) []Frame {
	totalFrames := 28
	frames := make([]Frame, 0, totalFrames+1)
	prevLines := 0

	for i := range totalFrames {
		var val, color string

		switch cfg.Mode {
		case model.RouletteStandard:
			n := game.RandomInt(0, 36)
			c := model.RouletteColors[n]
			color = colorForRoulette(c)
			val = fmt.Sprintf("%2d", n)
		case model.RouletteMinMax:
			n := game.RandomInt(cfg.Min, cfg.Max)
			val = fmt.Sprintf("%d", n)
			color = Cyan
		case model.RouletteCustomValues:
			val = game.RandomChoice(cfg.Values)
			color = Yellow
		}

		lines := buildRouletteBox(cfg.Mode, val, color, "", cfg)
		frames = append(frames, Frame{
			Content: redraw(lines, prevLines),
			Lines:   lines,
			Delay:   frameDelay(i),
		})
		prevLines = lineCount(lines)
	}

	// Final frame with the real result.
	var color, colorLabel string
	switch cfg.Mode {
	case model.RouletteStandard:
		color = colorForRoulette(result.Color)
		colorLabel = result.Color.String()
	case model.RouletteMinMax:
		color = Bold + Cyan
	case model.RouletteCustomValues:
		color = Bold + Yellow
	}

	lines := buildRouletteBox(cfg.Mode, result.Value, color, colorLabel, cfg)
	frames = append(frames, Frame{
		Content: redraw(lines, prevLines) + "\n",
		Lines:   lines,
		Delay:   0,
	})

	return frames
}

func buildRouletteBox(mode model.RouletteMode, value, color, colorLabel string, cfg model.RouletteConfig) []string {
	width := 17
	valPadded := centerPad(value, width-4)

	lines := []string{
		"+---------------+",
		"|     LUCK      |",
		"+---------------+",
		fmt.Sprintf("| %s%s%s |", color, valPadded, Reset),
	}

	switch mode {
	case model.RouletteStandard:
		label := centerPad("* "+colorLabel, width-4)
		lines = append(lines,
			"+---------------+",
			fmt.Sprintf("| %s%s%s |", color, label, Reset),
			"+---------------+",
		)
	case model.RouletteMinMax:
		rangeStr := fmt.Sprintf("[%d..%d]", cfg.Min, cfg.Max)
		label := centerPad(rangeStr, width-4)
		lines = append(lines,
			"+---------------+",
			fmt.Sprintf("| %s |", label),
			"+---------------+",
		)
	case model.RouletteCustomValues:
		lines = append(lines, "+---------------+")
	}

	return lines
}

// displayWidth returns the visible column width of a string,
// accounting for zero-width ANSI escape sequences.
func displayWidth(s string) int {
	// Strip ANSI escape sequences.
	stripped := s
	for {
		idx := strings.Index(stripped, "\033[")
		if idx < 0 {
			break
		}
		end := idx + 2
		for end < len(stripped) && !((stripped[end] >= 'A' && stripped[end] <= 'Z') || (stripped[end] >= 'a' && stripped[end] <= 'z')) {
			end++
		}
		if end < len(stripped) {
			end++
		}
		stripped = stripped[:idx] + stripped[end:]
	}
	return len([]rune(stripped))
}

func centerPad(s string, width int) string {
	dw := displayWidth(s)
	if dw >= width {
		return s
	}
	left := (width - dw) / 2
	right := width - dw - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// SlotsFrames generates ANSI animation frames for a slots spin.
func SlotsFrames(result model.SlotsResult, cfg model.SlotsConfig) []Frame {
	switch cfg.Mode {
	case model.SlotsStandard:
		return slotsStandardFrames(result, cfg)
	case model.SlotsMinMax:
		return slotsMinMaxFrames(result, cfg)
	case model.SlotsCustom:
		return slotsCustomFrames(result, cfg)
	default:
		return nil
	}
}

func slotsStandardFrames(result model.SlotsResult, cfg model.SlotsConfig) []Frame {
	var frames []Frame
	colWidth := 6
	prevLines := 0

	weights := make([]int, len(model.StandardSymbols))
	for i, s := range model.StandardSymbols {
		weights[i] = s.Weight
	}

	// The initial grid is the first cascade step's grid, or the final grid if no cascades.
	initialGrid := result.Grid
	if len(result.CascadeSteps) > 0 {
		initialGrid = result.CascadeSteps[0].Grid
	}

	// Phase 1: Reel-lock animation — columns lock left to right.
	totalSteps := cfg.Cols * 6
	lockedCols := 0

	for step := range totalSteps {
		col := step / 6
		if col > lockedCols {
			lockedCols = col
		}

		grid := make([][]string, cfg.Rows)
		for r := range cfg.Rows {
			grid[r] = make([]string, cfg.Cols)
			for c := range cfg.Cols {
				if c < lockedCols || (c == lockedCols && step%6 >= 5) {
					grid[r][c] = initialGrid[r][c]
				} else {
					idx := game.RandomWeighted(weights)
					grid[r][c] = model.StandardSymbols[idx].Symbol
				}
			}
		}

		lines := buildSlotsGridLines(grid, cfg.Cols, colWidth)
		lines = append(lines, buildStatusLine(cfg.Cols, colWidth, func(c int) bool {
			return c < lockedCols || (c == lockedCols && step%6 >= 5)
		}))

		delay := 80 * time.Millisecond
		if step%6 == 5 {
			delay = 500 * time.Millisecond
		}

		frames = append(frames, Frame{
			Content: redraw(lines, prevLines),
			Lines:   lines,
			Delay:   delay,
		})
		prevLines = lineCount(lines)
	}

	// Phase 2: Cascade animation — flicker wins, drop, repeat.
	for stepIdx, cs := range result.CascadeSteps {
		// Show the grid with all columns locked.
		lines := buildSlotsGridLines(cs.Grid, cfg.Cols, colWidth)
		lines = append(lines, buildStatusLine(cfg.Cols, colWidth, func(_ int) bool { return true }))
		lines = append(lines, fmt.Sprintf("%s%sCascade #%d — wins found!%s", Bold, Yellow, stepIdx+1, Reset))
		frames = append(frames, Frame{
			Content: redraw(lines, prevLines),
			Lines:   lines,
			Delay:   800 * time.Millisecond,
		})
		prevLines = lineCount(lines)

		// Flicker: alternate between showing symbols and "--" 3 times.
		for flick := range 6 {
			flickGrid := make([][]string, cfg.Rows)
			for r := range cfg.Rows {
				flickGrid[r] = make([]string, cfg.Cols)
				for c := range cfg.Cols {
					pos := model.Position{Row: r, Col: c}
					if cs.Removed[pos] {
						if flick%2 == 0 {
							flickGrid[r][c] = "--"
						} else {
							flickGrid[r][c] = cs.Grid[r][c]
						}
					} else {
						flickGrid[r][c] = cs.Grid[r][c]
					}
				}
			}
			lines = buildSlotsGridLines(flickGrid, cfg.Cols, colWidth)
			lines = append(lines, buildStatusLine(cfg.Cols, colWidth, func(_ int) bool { return true }))

			// Show which symbols won during the flicker.
			for _, w := range cs.Wins {
				name := symbolName(w.Symbol)
				lines = append(lines, fmt.Sprintf("  %s%s%s x%d on line %d",
					Cyan, name, Reset, w.Count, w.LineIndex+1))
			}

			frames = append(frames, Frame{
				Content: redraw(lines, prevLines),
				Lines:   lines,
				Delay:   150 * time.Millisecond,
			})
			prevLines = lineCount(lines)
		}

		// Show grid with removed positions as blanks (the "vanish").
		vanishGrid := make([][]string, cfg.Rows)
		for r := range cfg.Rows {
			vanishGrid[r] = make([]string, cfg.Cols)
			for c := range cfg.Cols {
				if cs.Removed[model.Position{Row: r, Col: c}] {
					vanishGrid[r][c] = "  "
				} else {
					vanishGrid[r][c] = cs.Grid[r][c]
				}
			}
		}
		lines = buildSlotsGridLines(vanishGrid, cfg.Cols, colWidth)
		lines = append(lines, buildStatusLine(cfg.Cols, colWidth, func(_ int) bool { return true }))
		lines = append(lines, "  Symbols removed...")
		frames = append(frames, Frame{
			Content: redraw(lines, prevLines),
			Lines:   lines,
			Delay:   600 * time.Millisecond,
		})
		prevLines = lineCount(lines)

		// Show the grid after drop + refill.
		lines = buildSlotsGridLines(cs.GridAfter, cfg.Cols, colWidth)
		lines = append(lines, buildStatusLine(cfg.Cols, colWidth, func(_ int) bool { return true }))
		lines = append(lines, fmt.Sprintf("  %sNew symbols dropped in! (%dx multiplier)%s",
			Green, stepIdx+2, Reset))
		frames = append(frames, Frame{
			Content: redraw(lines, prevLines),
			Lines:   lines,
			Delay:   800 * time.Millisecond,
		})
		prevLines = lineCount(lines)
	}

	// Phase 3: Final result summary.
	lines := buildSlotsGridLines(result.Grid, cfg.Cols, colWidth)
	lines = append(lines, buildStatusLine(cfg.Cols, colWidth, func(_ int) bool { return true }))

	if len(result.Paylines) > 0 {
		lines = append(lines, fmt.Sprintf("%s%s WIN! %d payline(s), %dx multiplier%s",
			Bold, Yellow, len(result.Paylines), result.Multiplier, Reset))
		for _, pl := range result.Paylines {
			lines = append(lines, fmt.Sprintf("  %s%s%s x%d on line %d",
				Cyan, symbolName(pl.Symbol), Reset, pl.Count, pl.LineIndex+1))
		}
		if result.Multiplier > 1 {
			lines = append(lines, fmt.Sprintf("  %sCascade bonus: %dx%s",
				Yellow, result.Multiplier, Reset))
		}
	} else {
		lines = append(lines, "No win this time.")
	}
	if result.FreeSpins > 0 {
		lines = append(lines, fmt.Sprintf("%s%s FREE SPINS x%d!%s", Bold, Magenta, result.FreeSpins, Reset))
	}
	if result.BonusRound {
		lines = append(lines, fmt.Sprintf("%s%s BONUS ROUND!%s", Bold, Red, Reset))
	}

	frames = append(frames, Frame{
		Content: redraw(lines, prevLines) + "\n",
		Lines:   lines,
		Delay:   0,
	})

	return frames
}

func symbolName(sym string) string {
	for _, s := range model.StandardSymbols {
		if s.Symbol == sym {
			return s.Name
		}
	}
	return sym
}

// buildStatusLine creates an aligned status row under a grid.
// Each cell is colWidth wide with a │ separator between them.
func buildStatusLine(cols, colWidth int, locked func(int) bool) string {
	var b strings.Builder
	b.WriteString(" ") // offset for leading │
	for c := range cols {
		if c > 0 {
			b.WriteString(" ") // offset for │ separator
		}
		if locked(c) {
			b.WriteString(centerPad("OK", colWidth))
		} else {
			b.WriteString(centerPad("...", colWidth))
		}
	}
	return b.String()
}

func buildSlotsGridLines(grid [][]string, cols, colWidth int) []string {
	rows := len(grid)
	var lines []string

	lines = append(lines, buildGridBorder(cols, colWidth))

	for r := range rows {
		if r > 0 {
			lines = append(lines, buildGridBorder(cols, colWidth))
		}
		var row strings.Builder
		row.WriteString("|")
		for c := range cols {
			padded := centerPad(grid[r][c], colWidth-1)
			row.WriteString(" " + padded + "|")
		}
		lines = append(lines, row.String())
	}

	lines = append(lines, buildGridBorder(cols, colWidth))
	return lines
}

func buildGridBorder(cols, colWidth int) string {
	var b strings.Builder
	b.WriteString("+")
	for c := range cols {
		if c > 0 {
			b.WriteString("+")
		}
		b.WriteString(strings.Repeat("-", colWidth))
	}
	b.WriteString("+")
	return b.String()
}

func slotsMinMaxFrames(result model.SlotsResult, cfg model.SlotsConfig) []Frame {
	var frames []Frame
	totalFrames := 25
	cols := cfg.Cols
	colWidth := 8
	prevLines := 0

	for i := range totalFrames {
		row := make([]string, cols)
		for c := range cols {
			row[c] = fmt.Sprintf("%d", game.RandomInt(cfg.Min, cfg.Max))
		}

		lines := buildSimpleGridLines([][]string{row}, cols, colWidth)
		lines = append(lines, "  ...")

		frames = append(frames, Frame{
			Content: redraw(lines, prevLines),
			Lines:   lines,
			Delay:   frameDelay(i),
		})
		prevLines = lineCount(lines)
	}

	// Final frame.
	coloredRow := make([]string, cols)
	for c := range cols {
		coloredRow[c] = Bold + Cyan + result.Grid[0][c] + Reset
	}
	lines := buildSimpleGridLines([][]string{coloredRow}, cols, colWidth)

	op := "+"
	if cfg.Operation == model.SlotsMultiply {
		op = "×"
	}
	var resultLine strings.Builder
	resultLine.WriteString("  ")
	for c := range cols {
		if c > 0 {
			resultLine.WriteString(" " + op + " ")
		}
		resultLine.WriteString(result.Grid[0][c])
	}
	fmt.Fprintf(&resultLine, " = %s%d%s", Bold+Green, result.FinalNumber, Reset)
	lines = append(lines, resultLine.String())

	frames = append(frames, Frame{
		Content: redraw(lines, prevLines) + "\n",
		Lines:   lines,
		Delay:   0,
	})

	return frames
}

func slotsCustomFrames(result model.SlotsResult, cfg model.SlotsConfig) []Frame {
	var frames []Frame
	totalFrames := 25
	cols := len(cfg.ReelValues)
	colWidth := 10
	prevLines := 0

	for i := range totalFrames {
		row := make([]string, cols)
		for c := range cols {
			row[c] = game.RandomChoice(cfg.ReelValues[c])
		}

		lines := buildSimpleGridLines([][]string{row}, cols, colWidth)

		frames = append(frames, Frame{
			Content: redraw(lines, prevLines),
			Lines:   lines,
			Delay:   frameDelay(i),
		})
		prevLines = lineCount(lines)
	}

	// Final frame.
	coloredRow := make([]string, cols)
	for c := range cols {
		coloredRow[c] = Bold + Yellow + result.Grid[0][c] + Reset
	}
	lines := buildSimpleGridLines([][]string{coloredRow}, cols, colWidth)

	frames = append(frames, Frame{
		Content: redraw(lines, prevLines) + "\n",
		Lines:   lines,
		Delay:   0,
	})

	return frames
}

func buildSimpleGridLines(grid [][]string, cols, colWidth int) []string {
	var lines []string

	lines = append(lines, buildGridBorder(cols, colWidth))

	for _, row := range grid {
		var r strings.Builder
		r.WriteString("|")
		for c := range cols {
			val := ""
			if c < len(row) {
				val = row[c]
			}
			padded := centerPad(val, colWidth-1)
			r.WriteString(" " + padded + "|")
		}
		lines = append(lines, r.String())
	}

	lines = append(lines, buildGridBorder(cols, colWidth))
	return lines
}

// CoinFlipFrames generates ANSI animation frames for a coin flip.
func CoinFlipFrames(result model.CoinFlipResult) []Frame {
	sides := []string{"(O)", " | ", "(O)", " | "}
	totalFrames := 20
	var frames []Frame
	prevLines := 0

	for i := range totalFrames {
		side := sides[i%len(sides)]
		lines := []string{
			"+---------------+",
			fmt.Sprintf("|  %s  |", centerPad(side, 11)),
			"+---------------+",
		}

		frames = append(frames, Frame{
			Content: redraw(lines, prevLines),
			Lines:   lines,
			Delay:   frameDelay(i),
		})
		prevLines = lineCount(lines)
	}

	// Final frame.
	inner := fmt.Sprintf("%s%s%s", Bold, result.Value, Reset)
	lines := []string{
		"+---------------+",
		fmt.Sprintf("|  %s  |", centerPad(inner, 11)),
		"+---------------+",
	}

	frames = append(frames, Frame{
		Content: redraw(lines, prevLines) + "\n",
		Lines:   lines,
		Delay:   0,
	})

	return frames
}

// DiceFrames generates ANSI animation frames for a dice roll.
func DiceFrames(result model.DiceResult) []Frame {
	totalFrames := 22
	var frames []Frame
	prevLines := 0

	for i := range totalFrames {
		dice := make([]int, result.Config.Count)
		for d := range result.Config.Count {
			dice[d] = game.RandomInt(1, result.Config.Sides)
		}

		lines := buildDiceLines(dice, result.Config.Sides, 0)

		frames = append(frames, Frame{
			Content: redraw(lines, prevLines),
			Lines:   lines,
			Delay:   frameDelay(i),
		})
		prevLines = lineCount(lines)
	}

	// Final frame.
	lines := buildDiceLines(result.Dice, result.Config.Sides, result.Sum)
	frames = append(frames, Frame{
		Content: redraw(lines, prevLines) + "\n",
		Lines:   lines,
		Delay:   0,
	})

	return frames
}

func buildDiceLines(dice []int, sides, sum int) []string {
	diceWidth := 7

	border := buildGridBorder(len(dice), diceWidth)

	var vals strings.Builder
	vals.WriteString("|")
	for d := range len(dice) {
		val := fmt.Sprintf("%d", dice[d])
		padded := centerPad(val, diceWidth)
		vals.WriteString(padded + "|")
	}

	lines := []string{border, vals.String(), border}

	diceLabel := fmt.Sprintf("%dd%d", len(dice), sides)
	if sum > 0 {
		lines = append(lines, fmt.Sprintf("  %s%s = %d%s", Bold+Green, diceLabel, sum, Reset))
	} else {
		lines = append(lines, "  ...")
	}

	return lines
}
