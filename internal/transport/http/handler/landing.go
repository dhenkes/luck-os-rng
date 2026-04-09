package handler

import (
	"fmt"
	"net/http"
	"strings"
)

type LandingHandler struct{}

func NewLandingHandler() *LandingHandler {
	return &LandingHandler{}
}

func (h *LandingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := getHost()

	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		h.serveHTML(w, host)
		return
	}
	h.serveText(w, host)
}

func (h *LandingHandler) serveText(w http.ResponseWriter, host string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, landingText, host, host, host, host, host, host, host, host, host, host, host, host)
}

func (h *LandingHandler) serveHTML(w http.ResponseWriter, host string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, landingHTML, host)
}

const landingText = `
  +----------------------------------------------+
  |         LUCK Unleashes Chaotic Kindness       |
  |       Self-hosted casino game suite & RNG     |
  +----------------------------------------------+

  Use curl -N for animated terminal output.

  ROULETTE
    curl -N %s/roulette
    curl -N "%s/roulette?mode=minmax&min=1&max=100"
    curl -N "%s/roulette?mode=custom&values=pizza,sushi,tacos,curry"

  SLOTS
    curl -N %s/slots
    curl -N "%s/slots?mode=minmax&min=1&max=100&cols=3&op=add"
    curl -N "%s/slots?mode=custom&reel1=pizza,sushi,tacos&reel2=fancy,casual&reel3=me,you,split"
    curl -N "%s/slots?luck=insane"

  COIN FLIP
    curl -N %s/coinflip
    curl -N "%s/coinflip?heads=yes&tails=no"

  DICE
    curl -N %s/dice
    curl -N "%s/dice?count=2&sides=6"

  DOUBLE OR NOTHING
    curl -N "%s/double?stake=100"

  ENGAGEMENT
    Add ?bet=medium|high|max for risk/reward multipliers.
    Add ?fast=1 for faster animations.
    Score, streak, and history are carried in the URL.

`

const landingHTML = `<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>LUCK — Casino Game Suite</title>
<style>
  body { margin: 8px; padding: 0; }
  pre { overflow-x: auto; font-family: 'Courier New', Courier, monospace; }
  @media (max-width: 480px) {
    pre { font-size: 11px; }
    h1 { font-size: 1.5em; }
  }
</style>
</head>
<body bgcolor="#c0c0c0">
<center>
<h1>LUCK</h1>
<h3><i>LUCK Unleashes Chaotic Kindness</i></h3>
<p>A self-hosted casino game suite &amp; configurable RNG</p>
<hr>
<p>
  <a href="/roulette">Roulette</a> |
  <a href="/slots">Slots</a> |
  <a href="/coinflip">Coin Flip</a> |
  <a href="/dice">Dice</a> |
  <a href="/double">Double</a>
</p>
<hr>
</center>

<h2>Games</h2>
<ul>
  <li><a href="/roulette">Roulette</a> &mdash; Spin the wheel. European 0-36, custom range, or pick from your own values.</li>
  <li><a href="/slots">Slots</a> &mdash; Pull the lever. Classic reels with paylines and cascading wins, or use as a multi-column randomizer.</li>
  <li><a href="/coinflip">Coin Flip</a> &mdash; Heads or tails. Custom labels supported.</li>
  <li><a href="/dice">Dice</a> &mdash; Roll any combination: d4, d6, d8, d10, d12, d20.</li>
  <li><a href="/double">Double or Nothing</a> &mdash; Risk your winnings on a coin flip.</li>
</ul>

<h2>Engagement</h2>
<p>Score, streak, and win history are carried in the URL -- no accounts, no database.</p>
<ul>
  <li><b>Bet tiers:</b> Add <tt>?bet=medium|high|max</tt> for risk/reward multipliers (3x, 10x, 100x).</li>
  <li><b>Streaks:</b> Build win streaks and climb the score leaderboard (of one).</li>
  <li><b>Double or Nothing:</b> After any win, gamble your points on a coin flip.</li>
  <li><b>Near-miss:</b> Watch for "SO CLOSE!" messages when you almost hit big.</li>
  <li><b>Speed:</b> Add <tt>?fast=1</tt> for faster animations.</li>
</ul>

<h2>curl commands</h2>
<p>Use <tt>curl -N</tt> for animated terminal output.</p>
<pre id="cmds"></pre>
<hr>
<p><font size="2"><i>Powered by crypto/rand. No accounts. No tracking. Just vibes.</i></font></p>
<script>
var h = '%s';
document.getElementById('cmds').innerHTML =
  '  <b>Roulette</b>\n' +
  '    curl -N ' + h + '/roulette\n' +
  '    curl -N "' + h + '/roulette?mode=minmax&min=1&max=100"\n' +
  '    curl -N "' + h + '/roulette?mode=custom&values=pizza,sushi,tacos,curry"\n\n' +
  '  <b>Slots</b>\n' +
  '    curl -N ' + h + '/slots\n' +
  '    curl -N "' + h + '/slots?mode=minmax&min=1&max=100&cols=3&op=add"\n' +
  '    curl -N "' + h + '/slots?mode=custom&reel1=pizza,sushi,tacos&reel2=fancy,casual&reel3=me,you,split"\n' +
  '    curl -N "' + h + '/slots?luck=insane"\n\n' +
  '  <b>Coin Flip</b>\n' +
  '    curl -N ' + h + '/coinflip\n' +
  '    curl -N "' + h + '/coinflip?heads=yes&tails=no"\n\n' +
  '  <b>Dice</b>\n' +
  '    curl -N ' + h + '/dice\n' +
  '    curl -N "' + h + '/dice?count=2&sides=6"\n' +
  '    curl -N "' + h + '/dice?count=1&sides=20"\n\n' +
  '  <b>Double or Nothing</b>\n' +
  '    curl -N "' + h + '/double?stake=100"\n\n' +
  '  <b>Bet Tiers</b>\n' +
  '    curl -N "' + h + '/slots?bet=high"\n' +
  '    curl -N "' + h + '/roulette?bet=max"';
</script>
</body>
</html>`
