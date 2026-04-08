package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/dhenkes/luck-os-rng/internal/game"
	"github.com/dhenkes/luck-os-rng/internal/model"
	"github.com/dhenkes/luck-os-rng/internal/renderer"
)

type RouletteHandler struct{}

func NewRouletteHandler() *RouletteHandler {
	return &RouletteHandler{}
}

func (h *RouletteHandler) Register(r chi.Router) {
	r.Get("/roulette", h.spin)
}

func (h *RouletteHandler) spin(w http.ResponseWriter, r *http.Request) {
	cfg, err := parseRouletteConfig(r)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := cfg.Validate(); err != nil {
		writeError(w, err)
		return
	}

	result := game.SpinRoulette(cfg)
	frames := renderer.RouletteFrames(result, cfg)
	streamOrPage(w, r, "Roulette", "/roulette", rouletteForm, frames)
}

func parseRouletteConfig(r *http.Request) (model.RouletteConfig, error) {
	q := r.URL.Query()

	mode, err := model.ParseRouletteMode(q.Get("mode"))
	if err != nil {
		return model.RouletteConfig{}, model.NewInvalidArgument(err.Error())
	}

	cfg := model.RouletteConfig{Mode: mode}

	switch mode {
	case model.RouletteMinMax:
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
	case model.RouletteCustomValues:
		raw := q.Get("values")
		if raw == "" {
			return cfg, model.NewInvalidArgument("values parameter is required for custom mode")
		}
		cfg.Values = strings.Split(raw, ",")
	}

	return cfg, nil
}

const rouletteForm = `
<h2>Roulette</h2>
<form id="cfg">
  <div class="field"><b>Mode:</b>
    <select name="mode" data-default="standard">
      <option value="standard">Standard (0-36)</option>
      <option value="minmax">Min/Max</option>
      <option value="custom">Custom Values</option>
    </select>
  </div>
  <div class="field"><b>Min:</b> <input name="min" type="number" value="" data-default="" placeholder="for minmax"></div>
  <div class="field"><b>Max:</b> <input name="max" type="number" value="" data-default="" placeholder="for minmax"></div>
  <div class="field"><b>Values:</b> <input name="values" type="text" value="" data-default="" placeholder="pizza,sushi,tacos (for custom)"></div>
</form>
`
