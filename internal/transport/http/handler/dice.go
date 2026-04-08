package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/dhenkes/luck-os-rng/internal/game"
	"github.com/dhenkes/luck-os-rng/internal/model"
	"github.com/dhenkes/luck-os-rng/internal/renderer"
)

type DiceHandler struct{}

func NewDiceHandler() *DiceHandler {
	return &DiceHandler{}
}

func (h *DiceHandler) Register(r chi.Router) {
	r.Get("/dice", h.roll)
}

func (h *DiceHandler) roll(w http.ResponseWriter, r *http.Request) {
	cfg := model.DiceConfig{}
	q := r.URL.Query()

	if v := q.Get("count"); v != "" {
		count, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, model.NewInvalidArgument("count must be an integer"))
			return
		}
		cfg.Count = count
	}
	if v := q.Get("sides"); v != "" {
		sides, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, model.NewInvalidArgument("sides must be an integer"))
			return
		}
		cfg.Sides = sides
	}

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		writeError(w, err)
		return
	}

	result := game.RollDice(cfg)
	frames := renderer.DiceFrames(result)
	streamOrPage(w, r, "Dice", "/dice", diceForm, frames)
}

const diceForm = `
<h2>Dice</h2>
<form id="cfg">
  <div class="field"><b>Count:</b> <input name="count" type="number" value="1" data-default="1" min="1" max="10"></div>
  <div class="field"><b>Sides:</b>
    <select name="sides" data-default="6">
      <option value="4">d4</option>
      <option value="6" selected>d6</option>
      <option value="8">d8</option>
      <option value="10">d10</option>
      <option value="12">d12</option>
      <option value="20">d20</option>
    </select>
  </div>
</form>
`
