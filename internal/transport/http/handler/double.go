package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/dhenkes/luck-os-rng/internal/game"
	"github.com/dhenkes/luck-os-rng/internal/model"
	"github.com/dhenkes/luck-os-rng/internal/renderer"
)

type DoubleHandler struct{}

func NewDoubleHandler() *DoubleHandler {
	return &DoubleHandler{}
}

func (h *DoubleHandler) Register(r chi.Router) {
	r.Get("/double", h.play)
}

func (h *DoubleHandler) play(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	stake, err := strconv.Atoi(q.Get("stake"))
	if err != nil || stake <= 0 {
		stake = 100
	}

	eng := parseEngagementState(r)
	host := getHost()

	// Cash-out mode: show celebration instead of flipping.
	if q.Get("cashout") == "1" {
		eng.Score += stake
		gameURLs := map[string]string{
			"Slots":     buildNextURL(host, "/slots", "", eng),
			"Roulette":  buildNextURL(host, "/roulette", "", eng),
			"Coin Flip": buildNextURL(host, "/coinflip", "", eng),
			"Dice":      buildNextURL(host, "/dice", "", eng),
		}
		frames := renderer.CashOutFrames(stake, eng, gameURLs)
		streamOrPage(w, r, "Cashed Out!", "/double", doubleForm, frames)
		return
	}

	cfg := model.CoinFlipConfig{Heads: "WIN", Tails: "LOSE"}
	result := game.FlipCoin(cfg)

	won := result.IsHeads

	var frames []renderer.Frame

	flipFrames := renderer.CoinFlipFrames(result)
	frames = append(frames, flipFrames...)

	prevLines := 0
	if len(frames) > 0 {
		prevLines = len(frames[len(frames)-1].Lines)
	}

	if won {
		newStake := stake * 2
		newEng := eng
		newEng.Score += stake
		newEng.Streak++
		newEng.History = game.AppendHistory(newEng.History, 'W')

		cashOutURL := fmt.Sprintf("%s/double?cashout=1&stake=%d&%s", host, newStake, newEng.QueryString())
		doubleAgainURL := buildDoubleURL(host, newStake, newEng)

		frames = append(frames, renderer.DoubleOrNothingFrame(true, stake, newStake, cashOutURL, doubleAgainURL, nil, prevLines))
	} else {
		newEng := eng
		newEng.Streak = 0
		newEng.History = game.AppendHistory(newEng.History, 'L')

		nextURL := buildDoubleURL(host, 100, newEng)
		gameURLs := map[string]string{
			"Slots":     buildNextURL(host, "/slots", "", newEng),
			"Roulette":  buildNextURL(host, "/roulette", "", newEng),
			"Coin Flip": buildNextURL(host, "/coinflip", "", newEng),
			"Dice":      buildNextURL(host, "/dice", "", newEng),
		}

		frames = append(frames, renderer.DoubleOrNothingFrame(false, stake, 0, nextURL, "", gameURLs, prevLines))
	}

	streamOrPage(w, r, "Double or Nothing", "/double", doubleForm, frames)
}

const doubleForm = `
<h2>Double or Nothing</h2>
<p>Risk your winnings on a coin flip. Win = double your stake. Lose = nothing.</p>
<form id="cfg">
  <div class="field"><b>Stake:</b> <input name="stake" type="number" value="100" data-default="100" min="1"></div>
</form>
`
