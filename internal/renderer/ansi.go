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
	Content string            // ANSI-encoded content for terminals.
	Lines   []string          // Raw display lines (with ANSI color but no cursor movement).
	Delay   time.Duration
	Tag     string            // Optional tag for special handling (e.g., "engagement").
	Meta    map[string]string // Optional metadata (e.g., URLs for browser buttons).
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
	if cfg.Mode == model.RouletteStandard {
		return rouletteWheelFrames(result)
	}
	return rouletteSimpleFrames(result, cfg)
}

// rouletteWheelFrames shows a spinning wheel strip for standard roulette.
func rouletteWheelFrames(result model.RouletteResult) []Frame {
	wheel := model.WheelOrder
	totalFrames := 28
	frames := make([]Frame, 0, totalFrames+1)
	prevLines := 0

	// Find final position in wheel.
	finalPos := 0
	for i, n := range wheel {
		if n == result.Number {
			finalPos = i
			break
		}
	}

	// Animate: ball moving around wheel, decelerating.
	startPos := game.RandomInt(0, len(wheel)-1)

	for i := range totalFrames {
		progress := float64(i) / float64(totalFrames)
		pos := startPos + int(float64(finalPos-startPos+len(wheel)*3)*progress)
		pos = ((pos % len(wheel)) + len(wheel)) % len(wheel)

		lines := buildWheelStrip(wheel, pos, false)
		frames = append(frames, Frame{
			Content: redraw(lines, prevLines),
			Lines:   lines,
			Delay:   frameDelay(i),
		})
		prevLines = lineCount(lines)
	}

	// Final frame with result highlighted.
	lines := buildWheelStrip(wheel, finalPos, true)

	frames = append(frames, Frame{
		Content: redraw(lines, prevLines) + "\n",
		Lines:   lines,
		Delay:   0,
	})

	return frames
}

// buildWheelStrip shows a horizontal strip of wheel numbers with the center highlighted.
func buildWheelStrip(wheel []int, centerPos int, highlight bool) []string {
	n := len(wheel)

	centerNum := wheel[centerPos]
	leftNum := wheel[((centerPos-1)%n+n)%n]
	rightNum := wheel[((centerPos+1)%n+n)%n]

	centerColor := colorForRoulette(model.RouletteColors[centerNum])
	leftColor := colorForRoulette(model.RouletteColors[leftNum])
	rightColor := colorForRoulette(model.RouletteColors[rightNum])

	// Each number is 2 digits. Center has brackets when highlighted.
	left := fmt.Sprintf("%s%02d%s", leftColor, leftNum, Reset)
	right := fmt.Sprintf("%s%02d%s", rightColor, rightNum, Reset)
	var center string
	if highlight {
		center = fmt.Sprintf("%s%s[%02d]%s", Bold, centerColor, centerNum, Reset)
	} else {
		center = fmt.Sprintf("%s%s %02d %s", Bold, centerColor, centerNum, Reset)
	}

	// Build value line with display-width-aware centering.
	// Inner content: "LL  CC  RR" (display width 12), padded to 15 for the box.
	inner := left + "  " + center + "  " + right
	padded := centerPad(inner, 15)
	valLine := "|" + padded + "|"

	// Position vv/^^ pointers at the center number.
	// leftPad = leading spaces from centerPad, then skip left number + separator.
	leftPad := (15 - displayWidth(inner)) / 2
	pointerOffset := 1 + leftPad + 2 + 2 + 1 // "|" + pad + left(2) + sep(2) + center leading char
	pointer := strings.Repeat(" ", pointerOffset) + "vv"

	lines := []string{
		"+---------------+",
		"|     LUCK      |",
		"+---------------+",
		pointer,
		valLine,
		strings.Repeat(" ", pointerOffset) + "^^",
		"+---------------+",
	}

	if highlight {
		colorLabel := model.RouletteColors[centerNum].String()
		labelPad := centerPad(colorLabel, 13)
		lines = append(lines,
			fmt.Sprintf("| %s%s%s |", centerColor, labelPad, Reset),
			"+---------------+",
		)
	}

	return lines
}

// rouletteSimpleFrames handles minmax and custom modes with the classic box.
func rouletteSimpleFrames(result model.RouletteResult, cfg model.RouletteConfig) []Frame {
	totalFrames := 28
	frames := make([]Frame, 0, totalFrames+1)
	prevLines := 0

	for i := range totalFrames {
		var val, color string
		switch cfg.Mode {
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

	var color string
	switch cfg.Mode {
	case model.RouletteMinMax:
		color = Bold + Cyan
	case model.RouletteCustomValues:
		color = Bold + Yellow
	}

	lines := buildRouletteBox(cfg.Mode, result.Value, color, "", cfg)
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
	totalFrames := 20
	var frames []Frame
	prevLines := 0

	// Two coin faces for animation. All padded to 7 lines to prevent flicker.
	face1 := []string{
		"      .----.",
		"     / o  o \\",
		"    |  \\__/  |",
		"     \\      /",
		"      '----'",
		"",
		"",
	}
	face2 := []string{
		"      .----.",
		"     /  ||  \\",
		"    |   ||   |",
		"     \\  ||  /",
		"      '----'",
		"",
		"",
	}
	edge := []string{
		"",
		"       ||||",
		"       ||||",
		"       ||||",
		"",
		"",
		"",
	}
	coinFaces := [][]string{face1, edge, face2, edge}

	for i := range totalFrames {
		face := coinFaces[i%len(coinFaces)]
		frames = append(frames, Frame{
			Content: redraw(face, prevLines),
			Lines:   face,
			Delay:   frameDelay(i),
		})
		prevLines = lineCount(face)
	}

	// Final frame: show result with label.
	var finalFace []string
	if result.IsHeads {
		finalFace = []string{
			fmt.Sprintf("      %s%s.----.%s", Bold, Yellow, Reset),
			fmt.Sprintf("     %s%s/ o  o \\%s", Bold, Yellow, Reset),
			fmt.Sprintf("    %s%s|  \\__/  |%s", Bold, Yellow, Reset),
			fmt.Sprintf("     %s%s\\      /%s", Bold, Yellow, Reset),
			fmt.Sprintf("      %s%s'----'%s", Bold, Yellow, Reset),
			"",
			fmt.Sprintf("    %s%s%s%s", Bold, Green, result.Value, Reset),
		}
	} else {
		finalFace = []string{
			fmt.Sprintf("      %s%s.----.%s", Bold, Cyan, Reset),
			fmt.Sprintf("     %s%s/  ||  \\%s", Bold, Cyan, Reset),
			fmt.Sprintf("    %s%s|   ||   |%s", Bold, Cyan, Reset),
			fmt.Sprintf("     %s%s\\  ||  /%s", Bold, Cyan, Reset),
			fmt.Sprintf("      %s%s'----'%s", Bold, Cyan, Reset),
			"",
			fmt.Sprintf("    %s%s%s%s", Bold, Red, result.Value, Reset),
		}
	}

	frames = append(frames, Frame{
		Content: redraw(finalFace, prevLines) + "\n",
		Lines:   finalFace,
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

// EngagementInfo holds the data needed to render the engagement footer.
type EngagementInfo struct {
	State      model.EngagementState
	IsWin      bool
	HasWinLoss bool     // true for slots (has win/loss), false for roulette/coin/dice
	Points     int
	NearMiss   []string
	NextURL    string
	DoubleURL  string
	ShareBlock []string // compact shareable result lines
}

// EngagementFrame builds a final frame showing score, streak, history, and next-play URLs.
// For terminals: appended after the game result as plain text.
// For browsers: sent as a special SSE event and rendered as HTML buttons.
func EngagementFrame(info EngagementInfo, prevLineCount int) Frame {
	lines := engagementLines(info)
	content := "\n"
	for _, line := range lines {
		content += line + "\n"
	}

	meta := map[string]string{"nextURL": info.NextURL}
	if info.HasWinLoss && info.IsWin && info.DoubleURL != "" {
		meta["doubleURL"] = info.DoubleURL
	}

	return Frame{
		Content: content,
		Lines:   lines,
		Delay:   0,
		Tag:     "engagement",
		Meta:    meta,
	}
}

func engagementLines(info EngagementInfo) []string {
	eng := info.State
	var lines []string

	lines = append(lines, "")

	// Score header.
	header := fmt.Sprintf("  %sScore: %d%s", Bold, eng.Score, Reset)
	if info.HasWinLoss {
		header += "  |  "
		if info.IsWin {
			header += fmt.Sprintf("%sStreak: W x%d%s", Green, eng.Streak, Reset)
		} else {
			header += fmt.Sprintf("%sStreak: ---%s", Red, Reset)
		}
	}
	if eng.Bet != model.BetLow {
		header += fmt.Sprintf("  |  BET: %s%s%s", Yellow, eng.Bet.Label(), Reset)
	}
	lines = append(lines, header)

	// Hot/cold streak display (only for games with win/loss).
	if info.HasWinLoss && eng.History != "" {
		histLine := renderHistory(eng.History)
		lines = append(lines, histLine)
	}

	// Points earned.
	if info.Points > 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  %s%s+%d points!%s", Bold, Green, info.Points, Reset))
	}

	// Near-miss messages.
	if !info.IsWin && len(info.NearMiss) > 0 {
		lines = append(lines, "")
		for _, msg := range info.NearMiss {
			lines = append(lines, fmt.Sprintf("  %s%s!! %s !!%s", Bold, Yellow, msg, Reset))
		}
	}

	// Share block.
	if len(info.ShareBlock) > 0 {
		lines = append(lines, "")
		lines = append(lines, info.ShareBlock...)
	}

	// Next play URLs (terminal gets curl, browser gets buttons via JS).
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %sGO AGAIN:%s", Bold, Reset))
	lines = append(lines, fmt.Sprintf("  curl -N \"%s\"", info.NextURL))

	if info.HasWinLoss && info.IsWin && info.DoubleURL != "" {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  %sDOUBLE OR NOTHING:%s", Bold, Reset))
		lines = append(lines, fmt.Sprintf("  curl -N \"%s\"", info.DoubleURL))
	}

	lines = append(lines, "")
	return lines
}

func renderHistory(history string) string {
	var b strings.Builder
	b.WriteString("  Recent: ")
	for i := 0; i < len(history); i++ {
		if i > 0 {
			b.WriteString(" ")
		}
		ch := history[i]
		if ch == 'W' {
			b.WriteString(Green + "W" + Reset)
		} else {
			b.WriteString(Red + "L" + Reset)
		}
	}

	// Detect and show streak.
	kind, count := detectHistoryStreak(history)
	if kind != "" {
		if kind == "HOT" {
			b.WriteString(fmt.Sprintf("  %s%s[HOT STREAK x%d]%s", Bold, Yellow, count, Reset))
		} else {
			b.WriteString(fmt.Sprintf("  %s%s[COLD STREAK x%d]%s", Bold, Cyan, count, Reset))
		}
	}

	return b.String()
}

func detectHistoryStreak(history string) (string, int) {
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

// ShareBlockSlots builds a compact shareable result for slots.
func ShareBlockSlots(result model.SlotsResult) []string {
	if result.Mode != model.SlotsStandard {
		return nil
	}
	var status string
	if len(result.Paylines) > 0 {
		status = fmt.Sprintf("WIN %dx %d line(s)", result.Multiplier, len(result.Paylines))
	} else {
		status = "no win"
	}
	row := ""
	if len(result.Grid) > 0 {
		for c, sym := range result.Grid[0] {
			if c > 0 {
				row += " "
			}
			row += sym
		}
	}
	return []string{
		"  +--- LUCK SLOTS ---+",
		fmt.Sprintf("  | %-18s |", row),
		fmt.Sprintf("  | %-18s |", status),
		"  +------------------+",
	}
}

// ShareBlockRoulette builds a compact shareable result for roulette.
func ShareBlockRoulette(result model.RouletteResult) []string {
	return []string{
		"  +--- LUCK ROULETTE ---+",
		fmt.Sprintf("  | %-20s |", result.Value+" "+result.Color.String()),
		"  +---------------------+",
	}
}

// ShareBlockCoinFlip builds a compact shareable result for coin flip.
func ShareBlockCoinFlip(result model.CoinFlipResult) []string {
	return []string{
		"  +--- LUCK COIN ---+",
		fmt.Sprintf("  | %-16s |", result.Value),
		"  +-----------------+",
	}
}

// ShareBlockDice builds a compact shareable result for dice.
func ShareBlockDice(result model.DiceResult) []string {
	diceStr := ""
	for i, d := range result.Dice {
		if i > 0 {
			diceStr += "+"
		}
		diceStr += fmt.Sprintf("%d", d)
	}
	return []string{
		"  +--- LUCK DICE ---+",
		fmt.Sprintf("  | %-16s |", fmt.Sprintf("%s = %d", diceStr, result.Sum)),
		"  +-----------------+",
	}
}

// SpinHeaderFrame shows a header for multi-spin mode.
func SpinHeaderFrame(current, total, runningScore int) Frame {
	lines := []string{
		"",
		fmt.Sprintf("  %s%s--- SPIN %d/%d --- (Score: %d) ---%s", Bold, Cyan, current, total, runningScore, Reset),
		"",
	}
	return Frame{
		Content: "\n" + lines[0] + "\n" + lines[1] + "\n" + lines[2] + "\n",
		Lines:   lines,
		Delay:   500 * time.Millisecond,
	}
}

// CashOutFrames renders a celebration for cashing out winnings.
func CashOutFrames(amount int, eng model.EngagementState, gameURLs map[string]string) []Frame {
	var frames []Frame

	// Jackpot-style celebration.
	colors := []string{Green, Yellow, Cyan}
	for i := 0; i < 3; i++ {
		c := colors[i%len(colors)]
		art := []string{
			"",
			fmt.Sprintf("  %s%s *  *  *  *  *  *  *  *  *  *  * %s", Bold, c, Reset),
			fmt.Sprintf("  %s%s*                                 *%s", Bold, c, Reset),
			fmt.Sprintf("  %s%s    C A S H E D   O U T !         %s", Bold, c, Reset),
			fmt.Sprintf("  %s%s*                                 *%s", Bold, c, Reset),
			fmt.Sprintf("  %s%s *  *  *  *  *  *  *  *  *  *  * %s", Bold, c, Reset),
			"",
		}
		prevLines := 0
		if i > 0 {
			prevLines = 7
		}
		frames = append(frames, Frame{
			Content: redraw(art, prevLines),
			Lines:   art,
			Delay:   300 * time.Millisecond,
		})
	}

	// Final frame with amount and game links.
	var lines []string
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %s%s+====================================+%s", Bold, Green, Reset))
	lines = append(lines, fmt.Sprintf("  %s%s|                                    |%s", Bold, Green, Reset))
	lines = append(lines, fmt.Sprintf("  %s%s|    YOU CASHED OUT: %-8d pts    |%s", Bold, Green, amount, Reset))
	lines = append(lines, fmt.Sprintf("  %s%s|    Total Score:   %-8d pts    |%s", Bold, Green, eng.Score, Reset))
	lines = append(lines, fmt.Sprintf("  %s%s|                                    |%s", Bold, Green, Reset))
	lines = append(lines, fmt.Sprintf("  %s%s+====================================+%s", Bold, Green, Reset))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %sPlay again:%s", Bold, Reset))

	for name, url := range gameURLs {
		lines = append(lines, fmt.Sprintf("    %s: curl -N \"%s\"", name, url))
	}
	lines = append(lines, "")

	// Pick first game URL for the web button.
	meta := make(map[string]string)
	for _, url := range gameURLs {
		meta["nextURL"] = url
		break
	}

	frames = append(frames, Frame{
		Content: redraw(lines, 7),
		Lines:   lines,
		Delay:   0,
		Tag:     "engagement",
		Meta:    meta,
	})

	return frames
}

// DoubleOrNothingFrame renders the double-or-nothing result and next options.
func DoubleOrNothingFrame(won bool, stake, newStake int, primaryURL, doubleURL string, gameURLs map[string]string, prevLineCount int) Frame {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %s%sDOUBLE OR NOTHING%s", Bold, Cyan, Reset))
	lines = append(lines, "")

	if won {
		lines = append(lines, fmt.Sprintf("  %s%sYOU WON! Stake doubled: %d -> %d%s", Bold, Green, stake, newStake, Reset))
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  %sCASH OUT:%s (keep your %d points)", Bold, Reset, newStake))
		lines = append(lines, fmt.Sprintf("  curl -N \"%s\"", primaryURL))
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  %sDOUBLE AGAIN:%s (risk %d for %d)", Bold, Reset, newStake, newStake*2))
		lines = append(lines, fmt.Sprintf("  curl -N \"%s\"", doubleURL))
	} else {
		lines = append(lines, fmt.Sprintf("  %s%sYOU LOST! %d points gone.%s", Bold, Red, stake, Reset))
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  %sTRY AGAIN:%s", Bold, Reset))
		lines = append(lines, fmt.Sprintf("  curl -N \"%s\"", primaryURL))
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  %sOTHER GAMES:%s", Bold, Reset))
		for name, url := range gameURLs {
			lines = append(lines, fmt.Sprintf("    %s: curl -N \"%s\"", name, url))
		}
	}

	lines = append(lines, "")

	content := "\n"
	for _, line := range lines {
		content += line + "\n"
	}

	meta := make(map[string]string)
	if won {
		meta["cashOutURL"] = primaryURL
		meta["doubleURL"] = doubleURL
	} else {
		meta["nextURL"] = primaryURL
	}

	return Frame{
		Content: content,
		Lines:   lines,
		Delay:   0,
		Tag:     "engagement",
		Meta:    meta,
	}
}

func buildDiceLines(dice []int, sides, sum int) []string {
	if sides == 6 {
		return buildD6Lines(dice, sum)
	}
	return buildNumericDiceLines(dice, sides, sum)
}

func buildNumericDiceLines(dice []int, sides, sum int) []string {
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

// d6Faces returns ASCII art for each d6 face value (1-6).
var d6Faces = map[int][]string{
	1: {
		"+-------+",
		"|       |",
		"|   o   |",
		"|       |",
		"+-------+",
	},
	2: {
		"+-------+",
		"| o     |",
		"|       |",
		"|     o |",
		"+-------+",
	},
	3: {
		"+-------+",
		"| o     |",
		"|   o   |",
		"|     o |",
		"+-------+",
	},
	4: {
		"+-------+",
		"| o   o |",
		"|       |",
		"| o   o |",
		"+-------+",
	},
	5: {
		"+-------+",
		"| o   o |",
		"|   o   |",
		"| o   o |",
		"+-------+",
	},
	6: {
		"+-------+",
		"| o   o |",
		"| o   o |",
		"| o   o |",
		"+-------+",
	},
}

func buildD6Lines(dice []int, sum int) []string {
	if len(dice) == 0 {
		return nil
	}

	// Build side-by-side dice faces.
	faceHeight := 5
	rows := make([]string, faceHeight)
	for d, val := range dice {
		face, ok := d6Faces[val]
		if !ok {
			face = d6Faces[1]
		}
		for r := 0; r < faceHeight; r++ {
			sep := ""
			if d > 0 {
				sep = "  "
			}
			rows[r] += sep + face[r]
		}
	}

	lines := rows

	diceLabel := fmt.Sprintf("%dd6", len(dice))
	if sum > 0 {
		lines = append(lines, fmt.Sprintf("  %s%s = %d%s", Bold+Green, diceLabel, sum, Reset))
	} else {
		lines = append(lines, "  ...")
	}

	return lines
}
