# LUCK

**LUCK Unleashes Chaotic Kindness** — a self-hosted casino game suite and configurable random number generator.

Single binary. No database. No accounts. Curl it from the terminal for animated ASCII art, or open it in a browser.

## Games

- **Roulette** — European wheel (0-36), custom min/max range, or pick from your own values
- **Slots** — Classic 5x3 reels with 20 paylines, cascading wins, wilds, scatters, and free spins. Also works as a multi-column randomizer (min/max or custom values)
- **Coin Flip** — Heads or tails with optional custom labels
- **Dice** — Roll 1-10 dice with d4, d6, d8, d10, d12, or d20

## Quick Start

```
go run cmd/server/main.go
```

Then:

```
curl -N localhost:8080/roulette
curl -N "localhost:8080/roulette?mode=custom&values=pizza,sushi,tacos,curry"
curl -N localhost:8080/slots
curl -N "localhost:8080/slots?luck=insane"
curl -N localhost:8080/coinflip
curl -N "localhost:8080/dice?count=2&sides=20"
```

Use `curl -N` (no-buffer) for real-time terminal animation.

Open `localhost:8080` in a browser for the web interface with config forms and curl command builder.

## Build

```
make build    # cross-compile for linux, macOS, windows
make test     # run tests
make cover    # test coverage
```

## Run

```
luck -addr :8080
```

The only option is `-addr` for the listen address.

### systemd

```ini
[Unit]
Description=LUCK
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/luck -addr :8123
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### nginx reverse proxy

```nginx
location / {
    proxy_pass http://127.0.0.1:8123/;
    proxy_buffering off;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

`proxy_buffering off` is important — without it nginx buffers the streaming animation.

## How Slots Work

Symbols: `CH` Cherry, `LM` Lemon, `OR` Orange, `GR` Grape, `BL` Bell, `DI` Diamond, `7s` Seven, `**` Wild, `$$` Bonus

- 3+ matching symbols left-to-right on any of 20 paylines = win
- Wild (`**`) substitutes for any symbol
- Winning symbols vanish, remaining drop down, new symbols fill the gaps (cascade)
- Each cascade round increases the multiplier (2x, 3x, 4x...)
- 3+ scatter (`$$`) anywhere = free spins

Use `?luck=high` for biased odds or `?luck=insane` for guaranteed cascades.

## License

AGPL-3.0 — see [LICENSE.md](LICENSE.md) for details. Commercial use requires permission.
