# Testing

Standard library only. No assertion library, no mocking framework, no golden
files, no `testdata/` directory. Tests are white-box: every `_test.go` declares
the same package as the code it tests and calls unexported functions directly.

Run them with the race detector, always:

```sh
go test -race ./...
```

## Shape

Table tests where there is more than one case, with `t.Run` so a failure names
itself:

```go
tests := []struct {
    name  string
    input string
    want  string
    wantOK bool
}{
    {name: "plain code", input: "др12", want: "dr12", wantOK: true},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, ok := PrepareCode(tt.input, formats)
        if got != tt.want || ok != tt.wantOK {
            t.Errorf("PrepareCode(%q) = (%q, %v), want (%q, %v)", tt.input, got, ok, tt.want, tt.wantOK)
        }
    })
}
```

Field names are `name`, the inputs, `want`, `wantOK`. Subtest names are
lowercase English phrases describing the case, not the input repeated.

Failure messages name the call, echo the input, then got, then want. A message
that prints only what happened forces the reader to open the test to find out
what was expected. `t.Fatal` inside a subtest, `t.Error` in a bare loop.

Helpers that fail on the caller's behalf start with `t.Helper()`. Helpers that
just build a value do not take `*testing.T` at all.

## Fixtures

Engine payloads are inline consts in the test file, copied from a real page,
with one line of comment saying what is wrong with them. That comment is the
point: the fixtures exist because the real data is malformed.

```go
// classicPayload is shaped like the real classic API answer: junk before the
// JSON, an unquoted null key, numbers arriving as strings.
const classicPayload = `jQuery163({"level": {` + ...
```

Shared builders (`classicGame()`, `testData()`, `testEnv(srv)`) live in
`helpers_test.go` in each package, not in whichever test file needed them
first.

## HTTP engines

Each engine test starts a `httptest.NewServer` and asserts on the request
inside the handler: method, path, query, cookie, form fields, encoding. The
handler is the assertion site; the test body checks the parsed result.

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    if got := r.PostForm.Get("cod"); got != "\xe4\xf012" {
        t.Errorf("cod = %q, want windows-1251 bytes", got)
    }
    w.Header().Set("Location", "?err=8")
    w.WriteHeader(http.StatusFound)
}))
defer srv.Close()

engine := newClassic(classicGame(), testEnv(srv))
```

`testEnv` points both engine base URLs at the test server. Status tables are
exercised by answering 302 with `Location: ?err=NN`. Cyrillic codes are
asserted as raw windows-1251 byte literals, because that is what the wire
carries.

Snapshot parsing can skip HTTP entirely by constructing the unexported snapshot
struct with the raw payload.

## What is worth testing here

Parsing and pure logic: `internal/game` sector and spoiler parsing,
`internal/htmltext` conversion and splitting, `internal/geo` coordinate
formats, `internal/updater` level detection, `internal/store` copying and
persistence, `internal/migrations` idempotence, menu render functions.

The thin Telegram wiring (`Ctx` reply helpers, command constructors, the
dispatch switch) is not tested and does not need to be; it has no logic worth
pinning and no way to fail quietly. Menu rendering, by contrast, is pure and is
covered, which is what catches a bad refactor of the shared render helper.
