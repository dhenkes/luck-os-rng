package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/dhenkes/luck-os-rng/internal/game"
	"github.com/dhenkes/luck-os-rng/internal/model"
	"github.com/dhenkes/luck-os-rng/internal/renderer"
)

type SlotsHandler struct{}

func NewSlotsHandler() *SlotsHandler {
	return &SlotsHandler{}
}

func (h *SlotsHandler) Register(r chi.Router) {
	r.Get("/slots", h.spin)
}

func (h *SlotsHandler) spin(w http.ResponseWriter, r *http.Request) {
	cfg, err := parseSlotsConfig(r)
	if err != nil {
		writeError(w, err)
		return
	}

	eng := parseEngagementState(r)
	cfg.Bet = eng.Bet

	if err := cfg.Validate(); err != nil {
		writeError(w, err)
		return
	}

	spins := cfg.Spins
	if spins < 1 {
		spins = 1
	}
	if spins > 10 {
		spins = 10
	}

	var frames []renderer.Frame
	var lastResult model.SlotsResult
	totalPoints := 0
	currentEng := eng

	for spin := range spins {
		result := game.SpinSlots(cfg)
		lastResult = result
		points, isWin := game.ScoreSlots(result, eng.Bet)
		totalPoints += points
		currentEng = game.UpdateEngagement(currentEng, isWin, points)

		spinFrames := renderer.SlotsFrames(result, cfg)

		if spins > 1 {
			header := renderer.SpinHeaderFrame(spin+1, spins, totalPoints)
			frames = append(frames, header)
		}

		frames = append(frames, spinFrames...)
	}

	_, lastIsWin := game.ScoreSlots(lastResult, eng.Bet)
	nearMiss := game.DetectSlotsNearMiss(lastResult, cfg)

	// Jackpot celebration animation (before engagement footer).
	jackpot := game.DetermineJackpotTier(lastResult, eng.Bet)
	if jackpot > model.JackpotNone {
		frames = append(frames, renderer.JackpotFrames(jackpot)...)
	}

	host := getHost()
	gameQS := gameQueryString(r)
	nextURL := buildNextURL(host, "/slots", gameQS, currentEng)
	doubleURL := buildDoubleURL(host, totalPoints, currentEng)

	prevLines := 0
	if len(frames) > 0 {
		prevLines = len(frames[len(frames)-1].Lines)
	}
	frames = append(frames, renderer.EngagementFrame(renderer.EngagementInfo{
		State:      currentEng,
		IsWin:      lastIsWin,
		HasWinLoss: true,
		Points:     totalPoints,
		NearMiss:   nearMiss,
		NextURL:    nextURL,
		DoubleURL:  doubleURL,
		ShareBlock: renderer.ShareBlockSlots(lastResult),
	}, prevLines))

	streamOrPage(w, r, "Slots", "/slots", slotsForm, frames)
}

func parseSlotsConfig(r *http.Request) (model.SlotsConfig, error) {
	q := r.URL.Query()

	mode, err := model.ParseSlotsMode(q.Get("mode"))
	if err != nil {
		return model.SlotsConfig{}, model.NewInvalidArgument(err.Error())
	}

	cfg := model.DefaultSlotsConfig()
	cfg.Mode = mode

	if v := q.Get("rows"); v != "" {
		rows, err := strconv.Atoi(v)
		if err != nil {
			return cfg, model.NewInvalidArgument("rows must be an integer")
		}
		cfg.Rows = rows
	}
	if v := q.Get("cols"); v != "" {
		cols, err := strconv.Atoi(v)
		if err != nil {
			return cfg, model.NewInvalidArgument("cols must be an integer")
		}
		cfg.Cols = cols
	}

	cfg.Luck = model.ParseSlotsLuck(q.Get("luck"))

	if v := q.Get("spins"); v != "" {
		spins, err := strconv.Atoi(v)
		if err != nil {
			return cfg, model.NewInvalidArgument("spins must be an integer")
		}
		cfg.Spins = spins
	}

	switch mode {
	case model.SlotsMinMax:
		min, err := strconv.Atoi(q.Get("min"))
		if err != nil {
			return cfg, model.NewInvalidArgument("min must be an integer")
		}
		max, err := strconv.Atoi(q.Get("max"))
		if err != nil {
			return cfg, model.NewInvalidArgument("max must be an integer")
		}
		cfg.Min = min
		cfg.Max = max
		cfg.Operation = model.ParseSlotsOperation(q.Get("op"))
		if cfg.Cols < 2 {
			cfg.Cols = 3
		}
	case model.SlotsCustom:
		var reelValues [][]string
		for i := 1; i <= 10; i++ {
			key := fmt.Sprintf("reel%d", i)
			raw := q.Get(key)
			if raw == "" {
				break
			}
			reelValues = append(reelValues, strings.Split(raw, ","))
		}
		if len(reelValues) == 0 {
			return cfg, model.NewInvalidArgument("at least one reel parameter required (reel1, reel2, ...)")
		}
		cfg.ReelValues = reelValues
		cfg.Cols = len(reelValues)
	}

	return cfg, nil
}

const slotsForm = `
<h2>Slots</h2>
<p>
  <b>Symbols:</b> CH=Cherry, LM=Lemon, OR=Orange, GR=Grape, BL=Bell, DI=Diamond, 7s=Seven, **=Wild, $$=Bonus<br>
  <b>How to win:</b> 3+ matching symbols left-to-right on any of 20 paylines. Wild (**) substitutes for anything.<br>
  <b>Cascades:</b> Happen automatically! When you win, those symbols vanish, remaining symbols drop down,
  new random symbols fill the gaps, and paylines are checked again. Each cascade round increases
  the multiplier (2x, 3x, 4x...). This repeats until no new wins form.<br>
  <b>Bonus:</b> 3+ scatter ($$) anywhere on the grid = 5 free spins. 4+ = bonus round.
</p>
<form id="cfg">
  <div class="field"><b>Mode:</b>
    <select name="mode" data-default="standard">
      <option value="standard">Standard</option>
      <option value="minmax">Min/Max</option>
      <option value="custom">Custom</option>
    </select>
  </div>
  <div class="field"><b>Rows:</b> <input name="rows" type="number" value="3" data-default="3"></div>
  <div class="field"><b>Cols:</b> <input name="cols" type="number" value="5" data-default="5"></div>
  <div class="field"><b>Luck:</b>
    <select name="luck" data-default="">
      <option value="">Normal</option>
      <option value="high">High (biased toward wins)</option>
      <option value="insane">Insane (guaranteed cascades)</option>
    </select>
  </div>
  <div class="field"><b>Bet:</b>
    <select name="bet" data-default="low">
      <option value="low">Low (1x)</option>
      <option value="medium">Medium (3x)</option>
      <option value="high">High (10x)</option>
      <option value="max">Max (100x)</option>
    </select>
  </div>
  <div class="field"><b>Spins:</b> <input name="spins" type="number" value="1" data-default="1" min="1" max="10"></div>
  <div class="field"><b>Min:</b> <input name="min" type="number" value="" data-default="" placeholder="for minmax"></div>
  <div class="field"><b>Max:</b> <input name="max" type="number" value="" data-default="" placeholder="for minmax"></div>
  <div class="field"><b>Op:</b>
    <select name="op" data-default="add"><option value="add">Add (+)</option><option value="multiply">Multiply (x)</option></select>
  </div>
  <div class="field"><b>Reel 1:</b> <input name="reel1" type="text" value="" data-default="" placeholder="for custom: a,b,c"></div>
  <div class="field"><b>Reel 2:</b> <input name="reel2" type="text" value="" data-default="" placeholder="x,y,z"></div>
  <div class="field"><b>Reel 3:</b> <input name="reel3" type="text" value="" data-default="" placeholder="1,2,3"></div>
</form>
`
