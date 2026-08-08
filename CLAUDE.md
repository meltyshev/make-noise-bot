# make-noise-bot

A Telegram bot for Dozor city quest games. One static binary, long polling, no
database: `config.json` and `state.json` sit in the working directory (or
wherever `--config` points).

```sh
go build .              # binary for your machine
go test -race ./...     # always -race, the bot has background goroutines
golangci-lint run       # the CI gate, see .golangci.yml
gofmt -w .
```

Install the linter with
`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2`.

## Layout

| Package | Role |
|---|---|
| `main.go` | flags, wiring, signal handling |
| `internal/bot` | Telegram dispatch, `Ctx`, commands, menus, callbacks |
| `internal/game` | Dozor engine clients: classic, lite, two prequels |
| `internal/updater` | poll loop, level-change detection, broadcasts |
| `internal/store` | in-memory state, atomic JSON persistence |
| `internal/migrations` | numbered raw-JSON state migrations |
| `internal/jsonfile` | reading and atomically replacing the JSON files |
| `internal/htmltext` | engine HTML to Telegram HTML, splitting, coordinate links |
| `internal/tgsend` | HTML delivery with fallbacks |
| `internal/texts` | every user-facing string |
| `internal/geo` | coordinate parsing, map service links |
| `internal/config` | `config.json` and the first-run wizard |
| `internal/secret` | redaction registry and a redacting slog handler |
| `internal/words`, `internal/avatar` | embedded dictionaries, PNG avatars |

Dependencies point one way and the graph is acyclic. `bot` and `updater` sit on
top, `game`, `tgsend`, `htmltext` and `config` in the middle, `store`, `geo`,
`texts`, `words`, `avatar`, `secret`, `jsonfile` and `migrations` at the bottom.
Keep it that way; if a new import would create a cycle, the design is wrong, not
the cycle.

## Hard rules

- ASCII punctuation only in code, comments and docs. No long dashes, no typographic
  quotes. Russian text in `internal/texts` is the exception, it is content.
- Every user-facing string lives in `internal/texts`. The one exception is the
  per-engine status tables in `internal/game`, which stay next to the status
  codes they map.
- Do not carry the old Python implementation into design decisions here. It
  lives on the `v1` branch and is not context for anything in this tree. The
  one pointer to it, in README.md, is for users and stays.
- No new dependencies without a reason that survives a day of thought. Four
  direct deps today.
- Commits and pushes are the maintainer's, never the agent's.

## Comments

A comment must add something the signature and body do not say in five seconds.
Nothing renders on pkg.go.dev, so completeness buys nothing here.

Write one for: package clauses (every package has one, keep it), identifiers that
cross a package boundary and carry a contract (`game.Engine`, `game.Snapshot`,
`store.View`/`Update`/`UpdateGame`, `secret.Redact`, `tgsend` splitting), and
anything whose reason is invisible - a workaround for a broken engine response,
a threshold, an ordering requirement.

Do not write one for one-line accessors, obvious constructors, or a restatement
of the next line.

Form: start with the identifier name, one full sentence, end with a period.
End-of-line comments may be fragments. Keep them short; two lines is usually
enough, five is a smell.

```go
// levelTask fingerprints what the team has to solve: the texts and the code
// layout, but not the codes they entered, the timer or the hints, which all
// change while the level stays the same.
func levelTask(snap game.Snapshot) string
```

## Naming

`MixedCaps`, initialisms keep their case (`chatID`, `URL`). No `Get` prefix on
getters. Receivers are one or two letters, the same letter for every method of
a type. Do not repeat the package name (`game.Engine`, not `game.GameEngine`).
Regex variables end in `Re` and live in one `var` block at the top of the file.
Never shadow a builtin - `new`, `max`, `close` and `len` are real functions.

## Errors

Wrap with `%w` at the end: `fmt.Errorf("open state: %w", err)`. Lowercase, no
trailing punctuation, no "failed" suffix. `errors.New` for static messages.
`error` is the last return value, always. Handle an error once: either return
it or log it, not both.

In `internal/bot` the funnel is `App.reportError`, which logs and DMs the admin
at most once per 30 seconds. Use it for failures the maintainer must know about.
Use `log.Warn` for the ordinary ones (a send to one chat failed, the engine
timed out); the same class of failure gets the same level everywhere.

Type assertions use the comma-ok form. Return an extra `ok bool` rather than a
sentinel value like `-1` or `""`.

## Logging

`log/slog` only, injected as a `*slog.Logger` field named `log`. Messages are
static lowercase noun phrases (`"engine load failed"`, `"bot started"`), never
formatted. Keys are snake_case, errors always under `"error"`.

## Context and concurrency

`ctx` is the first parameter of anything doing I/O. It is stored in a struct in
exactly two places, `Ctx` and `cb`, because the Telegram library hands handlers
no context; do not add a third. Every goroutine exits on `ctx.Done()` - there
are only two, the updater loop and the conversation janitor.

All persisted state goes through `store.View` and `store.Update`, which run the
callback under the mutex; do not retain the `*Data` past the callback. Values
returned from the store are copies, so mutating them is safe and pointless.

## Telegram

- Callback data is namespaced `ns:action:arg` and capped at 64 bytes. Live
  namespaces: `perm`, `cfg`, `ch`, `gc`, `gs`, `cs`, `res`, `gm`. The two the
  updater also needs are constants in `texts`.
- Menus show state with `mark()`, which prefixes a checkmark. One convention
  everywhere, including engine and code-format pickers.
- Config menus are private-chat only.
- Messages go out as HTML through `tgsend`, which splits at 4000 UTF-16 units,
  balances tags across parts and, if Telegram still rejects the markup, resends
  it as an escaped code block and only then as plain text. Never call
  `SendMessage` with HTML directly.
- Chat message prefixes: `!` and `.` send a code as typed, `?` prints the board,
  `$` answers a pinned level, `&` opens a spoiler.

## Engines

`internal/game` is reverse-engineered from live sites and encodes years of
knowledge: windows-1251 bodies, JSONP wrappers with invalid JSON, HTML marker
comments, `err=N` redirect status tables. The three engine implementations -
classic, lite and the one shared by both prequels - look alike on purpose and
are not to be merged.

Never change parsing without a fixture test built from a real page. Fixtures are
inline consts in the test file with one line saying what is broken about them.

## State and migrations

`state.json` and `config.json` are both written atomically through
`internal/jsonfile` (temp file, rename, mode 0600). Renaming, removing or
reinterpreting a field of `store.Data` needs a migration in
`internal/migrations`: a new `000N_name.go`, registered in `init()`, operating
on `map[string]any` rather than the current structs, plus a bump of the current
schema version. A purely additive `omitempty` field needs none, because an
absent key already unmarshals to the zero value the code expects. Migrations
must be idempotent on their own - the version gate in `Apply` is not the thing
that makes them safe - and must never lose a field they do not understand.

## Tests

White-box, in-package, standard library only. No testify, no go-cmp, no mocking
framework, no `t.Parallel()`.

Table tests with `t.Run` and fields named `name`, the inputs, `want`, `wantOK`.
Failure messages read `Func(input) = got, want want`. Helpers that fail on the
caller's behalf call `t.Helper()`. HTTP engines are tested against
`httptest.NewServer` with the assertions inside the handler and 302 `?err=NN`
answers for status tables. Use `t.Context()` and `t.TempDir()`.

## Commits

`feat:`, `fix:`, `ref:`, `docs:`, `test:`, `ci:` plus a capitalised imperative
subject: `feat: Send spoiler codes with the & prefix`. One concern per commit.

## More

- [docs/style.md](docs/style.md) - the long form, and what was deliberately rejected.
- [docs/architecture.md](docs/architecture.md) - how an update becomes a message.
- [docs/testing.md](docs/testing.md) - fixtures and test patterns in detail.
