# Make Noise Bot

A Telegram bot for playing [Dozor](https://dzzzr.ru) city quest games. It
watches the game engine, shouts "АП!" the moment a new level starts, forwards
questions, hints and solved spoilers to your team chats - and every message in
an allowed chat is treated as a potential game code and submitted to the
engine automatically.

One static binary, no database, no webserver. Config and state are two small
JSON files next to the executable.

> The previous Python/Heroku implementation lives in the
> [`v1`](https://github.com/meltyshev/make-noise-bot/tree/v1) branch.

## Quick start

1. Download the binary for your OS from
   [Releases](https://github.com/meltyshev/make-noise-bot/releases) (or see
   [Building](#building)).
2. Create a bot with [@BotFather](https://t.me/BotFather) and copy the token.
3. Run the binary. It asks for the token, checks it and writes `config.json`:

   ```
   $ ./make-noise-bot
   Файл конфигурации не найден - настроим бота.
   Вставьте токен бота: <paste>
   Готово, это @your_bot. Конфигурация сохранена в config.json.
   Теперь отправьте боту /start - первый написавший станет админом.
   ```

4. Send `/start` to the bot - the first person to do it becomes the admin.

That's it: the bot is polling. No ports, no domains, no databases.

### Non-interactive setup

```sh
./make-noise-bot --token 123456:ABC-DEF          # creates config.json
./make-noise-bot --token 123456:ABC-DEF --admin-id 111111
```

### Docker

```sh
docker compose run --rm make-noise-bot   # first run: interactive setup into ./data
docker compose up -d
```

## Playing a game

1. `/gameconfig` - pick the engine (DozorClassic / DozorLite / prequels) and
   fill in the city, credentials and code formats.
2. `/game` - start: the bot logs into the engine and answers with the game
   link. Run `/game` again after the finish to stop.
3. `/subscribe` in every chat that should receive updates, then pick what it
   gets: АП!, подсказки, спойлеры, задание, примечания. Any number of chats
   can receive any combination; the same lists are in `/gameconfig`.
4. Play. Type codes right into the chat - `др12`, `dr12`, `--12` are
   normalized by the configured formats; `!код` submits as-is. `?` prints the
   sector board and progress, `$код` answers a pinned (сквозной) level and
   `&код` opens a spoiler.

Chats must be allowed first: `/permission` sends the admin a request with
Allow/Forbid buttons (manual `/allow <id>` works too). Managers and
subscribers are picked from button lists in `/config` and `/gameconfig`.

## Commands

Utilities: `/morse`, `/anagram`, `/mask` (offline dictionaries built into the
binary), `/intersection`, `/letterstonumbers`, `/numberstoletters`,
`/coordinates`, `/help`, `/cancel`.

Game: `/question`, `/notes`, `/link`, `/rating`, `/clearrating`,
`/restrict`, `/bruteforce`, `/pinlevel`, `/unpinlevel`, `/subscribe`,
`/game`, `/gameconfig` (code formats are preset buttons there and apply to a
running game immediately).

Admin: `/permission`, `/allow`, `/forbid`, `/chats`, `/drop`, `/write`,
`/config` (managers, map service, leave mode), `/chatid`, `/userid`,
`/avatar`.

Coordinates in level texts, notes and hints become links to the map service
picked in `/config` (Яндекс.Карты, Google Maps, 2ГИС or OpenStreetMap), so
they open the map app on a phone.

## Configuration

`config.json` (created by the setup wizard, permissions `0600`):

```json
{
  "token": "123456:ABC-DEF",
  "admin_id": 111111,
  "update_interval_seconds": 5,
  "state_path": "state.json",
  "debug": false
}
```

- `update_interval_seconds` - how often the engine is polled during a game.
- `state_path` - where chats, permissions, ratings and the current game are
  stored. Back up this file if you care about ratings.
- `debug` - dump raw engine responses to `debug/` when parsing fails; also
  available as the `--debug` flag.

## Building

Go 1.26+:

```sh
go build .          # binary for your machine
go test ./...
```

Cross-compilation is a plain `GOOS`/`GOARCH` matter (CGO is not used), and
releases are produced by GoReleaser from a `v*` tag.

## Notes

- The engine scrapers in `internal/game` encode years of reverse engineering:
  windows-1251 encodings, malformed-JSON fixups and status-code tables. Treat
  changes there with respect and keep the tests green.
- The bot auto-renews the engine session mid-game if it expires.
- Embedded assets and their licenses are listed in [NOTICE](NOTICE).

## License

MIT - see [LICENSE](LICENSE).
