# Style, the long form

[CLAUDE.md](../CLAUDE.md) has the rules. This file has the reasoning, and the
list of widely recommended practices this project deliberately does not follow,
so they do not get re-litigated every few months.

Sources: [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments),
[Effective Go](https://go.dev/doc/effective_go),
[go.dev/doc/comment](https://go.dev/doc/comment),
[Google Go Style Guide](https://google.github.io/styleguide/go/),
[Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).

## Why the comment rule is what it is

Code Review Comments asks for doc comments on exported names "as should
non-trivial unexported type or function declarations". Google's guide narrows
it further, to declarations "with unobvious behavior". Effective Go says
comments "do not need to be elaborate, in fact they should not be elaborate".
None of them asks for a comment on every function.

The rule that requires one on every exported name exists to make pkg.go.dev
readable for strangers. Everything here is under `internal/`, so nothing is
importable from outside the module and nothing renders publicly. What survives
is the underlying reason: a reader of `game.Engine` is in a different package
and cannot see the implementation, so that contract needs words. A reader of
`Ctx.Text()` is three lines away from the body and needs none.

Both restic and prometheus reach the same conclusion in their linter configs
and switch the exported-comment enforcement off. This project does the same:
`revive`'s `exported` rule is disabled, `package-comments` stays on.

The format rule is not relaxed, because it costs nothing. When a comment is
written, exported or not, it starts with the identifier name, is a full
sentence, and ends with a period.

## Errors

Wrapping with `%w` is for callers that will use `errors.Is` or `errors.As`.
This project has almost none of those, so the honest reason to wrap is the
message: the chain of "load config", "migrate state", "open state", "start bot"
is what makes a one-line log entry diagnosable. Keep wrapping for that, keep the
annotation informative, and do not add one that only says "it failed".

Handle an error once. If a function returns an error, it should not also log
it. The exceptions are the two top-level loops, where there is nobody left to
return to: `App.reportError` and `Updater.report`.

`App.reportError` logs at error level and DMs the admin, rate limited to one
message per 30 seconds, with secrets redacted and the text truncated. It is for
"the maintainer needs to know". Everything routine is `log.Warn` and nothing
else, at the same level for the same class of failure.

## Interfaces

`game.Engine` and `game.Snapshot` are the only interfaces, and they live in the
package that implements them rather than in the consumers. That is the opposite
of the usual advice, and it is correct here: they are the product of the
package, polymorphism over four game variants, not a seam invented for mocking.
Google's guide allows exactly this case.

Do not add interfaces for `store.Store` or `*tgbot.Bot`. There is one
implementation of each and the tests that matter exercise pure functions.

## Deliberately not doing

| Practice | Why not |
|---|---|
| Doc comment on every exported name | Nothing is public, see above. |
| `testify`, `go-cmp` | Standard library assertions keep the dependency count at four. Assertion helpers are not idiomatic Go anyway. |
| `t.Parallel()` everywhere | The suite is pure CPU table tests and finishes in seconds. Parallel subtests buy nothing and invite capture bugs. |
| External `_test` packages | Tests call unexported helpers on purpose; forcing exports for testability is worse. |
| `var _ Engine = (*Classic)(nil)` | Each engine is assigned to `game.Engine` at a single call site, so the compiler already catches drift with a better message. |
| `_`-prefixed package globals | An Uber house rule that contradicts everyone else. |
| Copying slices and maps at every boundary | Applied where a mutex is involved, in `store`. Everywhere else it is allocation for a threat that does not exist inside one binary. |
| Line length, function length, cyclomatic complexity limits | Numeric thresholds arbitrate between team members. Go has no line length limit by design; break on meaning instead. |
| Wrapping every returned error (`wrapcheck`) | Wrapping is deliberate here, not mechanical. |
| Merging the three engine parsers | They read alike but differ in ways that only show up against a live site. They stay separate and separately tested. |
| Renaming callback data strings | They are a wire protocol between messages already sent and the running binary. Renaming them breaks buttons in old messages for no gain. |
| Moving the engine status tables into `texts` | The Russian strings there are one half of a status-code table. Split them and the table stops being readable. |

## The linter

`.golangci.yml` is the CI gate. It is deliberately a correctness set, not a
style set: things that catch bugs, plus the naming and error-shape rules the Go
guides actually prescribe. `gocritic` runs without its `style` tag because that
overlaps `revive`, and with `hugeParam` and `rangeValCopy` off because
`store.Game` and `store.GameConfig` are passed by value on purpose - an engine
gets a snapshot of the game, not a handle on the stored one.

`modernize` is on and it is the one that will rewrite code shape: `new(expr)`
instead of pointer helpers, `slices.Contains` instead of search loops,
`strings.Cut` instead of index arithmetic. Read its fixes before accepting
them; `golangci-lint run --fix` produces correct but occasionally ugly names.

Adding a `//nolint` requires the specific linter and an explanation, enforced
by `nolintlint`. If a rule fires often enough to be annoying, turn the rule off
in the config with a comment saying why, rather than sprinkling suppressions.
