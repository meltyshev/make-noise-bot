# Architecture

One process, two loops. The Telegram loop reacts to what people type; the
updater loop polls the game engine and pushes what changed. They share one
`store.Store` and never talk to each other directly.

```
main.go
  config.Load / config.Wizard        -> config.json
  migrations.Apply                   -> state.json, raw JSON
  store.Open                         -> state.json, typed
  bot.New(cfg, store, logger)
  go updater.New(...).Run(ctx)       loop 2
  app.Run(ctx)                       loop 1, blocks until SIGINT/SIGTERM
```

## Loop 1: an update becomes a reply

`App.onUpdate` (`internal/bot/dispatch.go`) is the only entry point. The
`go-telegram/bot` library is configured with `WithNotAsyncHandlers`, so updates
arrive one at a time and handlers never run concurrently with each other.

Order matters and is fixed:

1. Callback query -> `onCallback`, routed by the namespace in `callback_data`.
2. No admin configured yet -> the first `/start` in a private chat claims it.
   Nothing else is processed until then.
3. `/command` -> the registry in `App.commands`. Unknown commands are swallowed
   so a typo is never submitted to the engine as a code.
4. An open conversation -> `cmd.Handle` with the stored state. Conversations
   expire after an hour; a janitor goroutine sweeps them.
5. Leave mode in a group -> the bot leaves.
6. Otherwise the message is game input: `?` prints the board, `$` answers a
   pinned level, `&` opens a spoiler, anything else is a code.

`Ctx` (`internal/bot/ctx.go`) is the per-update handle: the message, the app,
the context, and the reply helpers. `cb` (`internal/bot/callbacks.go`) is its
equivalent for callback queries, with `answer` and `edit` instead of `Reply`.

A command is a `Command{Name, Description, Init, Handle}`. `Init` runs on
`/name`, `Handle` runs on the next message when `Init` opened a conversation.
Commands with an argument accept it inline (`/mask ко_ка`) or by asking.

## Loop 2: the engine becomes a broadcast

`Updater.tick` (`internal/updater/updater.go`) runs on a ticker, default every
5 seconds, and never overlaps itself. It stops early when there is no game or
nobody subscribed.

Each tick loads a `game.Snapshot` and compares it with what the store
remembers. State is persisted before anything is sent, so a broadcast that
fails is never repeated.

**Level changes are decided by a vote.** Organizers renumber levels when they
delete one and rewrite the task when they fix a typo, so no single signal is
trustworthy. `isNewLevel` counts three: the level number changed, the task
fingerprint changed, the timer fell back into the first 120 seconds. Two of
three agree or nothing is announced.

The task fingerprint (`levelTask`) is a SHA-256 of the question plus the sector
layout. Entered codes, the timer and hints are excluded on purpose: they change
constantly while the level stays the same.

Hints and spoilers are diffed the same way, against `HintNumber` and
`SolvedSpoilers` in the stored game.

Every broadcast goes through `announce`, which sends each subscribed chat the
version it asked for: the full text, or a one-line notice for chats subscribed
to events only.

## Engines

`game.New(storedGame, env)` returns an `Engine` for the configured engine name.
Four exist: DozorClassic, DozorLite and a prequel of each. `Engine.Load`
returns a `Snapshot`, which is a read-only view of the level: number, question,
progress, sectors, hint, spoilers, time on level.

Sessions expire mid-game. Classic and its prequel detect it (an HTML login page
where JSON was expected), re-login, retry once, and hand the new session back
through `Env.OnSessionUpdate` so the bot can persist it.

Everything the engines produce is HTML written by game organizers, which means
it is frequently invalid. `internal/htmltext` converts it to the small tag set
Telegram accepts, balances tags, turns coordinates into map links, and splits
long messages without breaking markup. `internal/tgsend` sends the result and
falls back to plain text if Telegram still refuses it.

## State

`store.Data` holds managers, chats and permissions, the map service, user
names, the game config, the running game and the rating. `store.View` and
`store.Update` take a callback that runs under the mutex; `Update` persists on
every change through `internal/jsonfile`, which writes a temp file and renames
it over `state.json` so a crash mid-write cannot truncate the state.

Getters return copies, so a caller cannot reach into stored state by accident.

`internal/migrations` runs before the store opens, on raw `map[string]any`, so
a migration is not tied to the current struct definitions. Files are numbered
(`0001_...`, `0002_...`), register themselves in `init()`, and are stamped in
`schema_version`. A fresh file is stamped at the current version and no
migration runs.

## Secrets

The bot token and the game password are registered with `internal/secret` at
startup. A wrapping `slog.Handler` redacts them from every log record, and
transport errors have their URL stripped before they are logged or sent to the
admin, because Telegram puts the token in the request URL.
