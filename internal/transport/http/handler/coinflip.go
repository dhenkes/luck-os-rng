package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dhenkes/luck-os-rng/internal/game"
	"github.com/dhenkes/luck-os-rng/internal/model"
	"github.com/dhenkes/luck-os-rng/internal/renderer"
)


type CoinFlipHandler struct{}

func NewCoinFlipHandler() *CoinFlipHandler {
	return &CoinFlipHandler{}
}

func (h *CoinFlipHandler) Register(r chi.Router) {
	r.Get("/coinflip", h.flip)
}

func (h *CoinFlipHandler) flip(w http.ResponseWriter, r *http.Request) {
	cfg := model.CoinFlipConfig{
		Heads: r.URL.Query().Get("heads"),
		Tails: r.URL.Query().Get("tails"),
	}
	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		writeError(w, err)
		return
	}

	result := game.FlipCoin(cfg)
	frames := renderer.CoinFlipFrames(result)
	streamOrPage(w, r, "Coin Flip", "/coinflip", coinflipForm, frames)
}

const coinflipForm = `
<h2>Coin Flip</h2>
<form id="cfg">
  <div class="field"><b>Heads:</b> <input name="heads" type="text" value="" data-default="" placeholder="Heads"></div>
  <div class="field"><b>Tails:</b> <input name="tails" type="text" value="" data-default="" placeholder="Tails"></div>
</form>
`
