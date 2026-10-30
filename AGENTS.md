# AGENTS.md — Pokedex CLI (Go)

An interactive Pokedex for the terminal, built on the [PokéAPI](https://pokeapi.co). Module `github.com/novotnyradekcz/pokedexcli`, Go 1.26.4. The only third-party dependency is `golang.org/x/term`; prefer the standard library.

## Commands
- run: `go run .`
- build: `go build ./...`
- test: `go test -race ./...` (add `-count=1` to bypass the test cache)
- single test: `go test -race -run TestName ./internal/repl`
- format check: `gofmt -l .` (must print nothing; fix with `gofmt -w .`)
- vet: `go vet ./...`
- tidy: `go mod tidy`

**Required checks before finishing any change:** `gofmt -l .` is empty, `go vet ./...` is clean, `go test -race ./...` passes. Do not add `golangci-lint` or `goimports` requirements or a `.golangci.yml`.

## Layout
| Path | Responsibility |
|---|---|
| `main.go` | Wiring only: create the client, call `repl.Run`. No logic. |
| `internal/pokeapi` | PokéAPI client and response types. All HTTP lives here. |
| `internal/pokecache` | Concurrent in-memory cache with expiry. |
| `internal/pokedex` | Caught-Pokémon collection; remembers catch order. |
| `internal/repl` | Command loop, terminal input, and the commands. |

## Project conventions
- **Output:** commands print only through `cfg.out` (`fmt.Fprintf(cfg.out, ...)`), never `fmt.Println`.
- **Randomness:** use `cfg.roll`, never `math/rand` directly in a command.
- **API access:** `repl` depends on the `repl.API` interface, not the concrete client.
- **No `os.Exit` in commands.** `exit` returns `errQuit`; the loop stops on it.
- **Caching:** keys are URLs. Cache a body only after it decoded successfully; never cache errors. The first location-area page is requested as `?offset=0&limit=20` so it shares a cache entry with the API's own `previous` URL (this makes `mapb` hit the cache).
- **Pokedex keys** are the API's Pokémon names, so `inspect` takes names, not ids.
- **User-visible strings** (for example `you're on the first page`, `pikachu was caught!`) follow the boot.dev course spec and are asserted by tests. Do not reword them unless asked.

### Adding a command
1. Create `internal/repl/command_<name>.go` with `func commandX(ctx context.Context, cfg *config, args ...string) error`. Validate `args` and return a `usage: ...` error.
2. Register it in `getCommands()` in `internal/repl/commands.go`; it then appears in `help`.
3. Test it in `internal/repl` using `newTestConfig(newFakeAPI())` and the `run` helper from `commands_test.go`.

## Errors
- Return errors; never discard them with `_` in production code. Accepted exceptions, each with a short comment: `defer resp.Body.Close()` on a read-only body, and `fmt.Fprint*` writes to `cfg.out`. For other deferred cleanup, report the error with a named return and `errors.Join` (see `terminalReader.ReadLine`).
- Wrap with `%w` when you add context: `fmt.Errorf("decoding response from %s: %w", url, err)`. Messages are lowercase, with no `error:` or `failed to` prefix. Don't wrap when the error already carries the context (`net/http` errors name the method and URL).
- Inspect errors with `errors.Is` and `errors.As`, never by comparing strings.
- Use a typed error when callers need data (`*pokeapi.StatusError`). Use a sentinel when identity is enough (`var errQuit = errors.New("quit")`). Exported sentinels are named `ErrXxx`.
- Handle an error once: do not log and return it. The REPL loop is the only place that prints command errors; commands return them.
- Errors reach the user as typed. Keep them short and readable. When translating to a friendly message, keep the cause reachable with `Unwrap` (see `notFoundError`).
- `panic` is only for programmer errors in constructors (for example `NewCache` with a non-positive interval), and it must be documented. Never panic on user input, network, or I/O failures.

## Context
- Functions that do network I/O or other blocking, cancellable work take `ctx context.Context` as the first parameter and pass it down. Create requests with `http.NewRequestWithContext`, never `http.Get`.
- Never store a `context.Context` in a struct field. Pass it as a parameter (command callbacks receive it).
- Call `defer cancel()` (or `defer stop()` for `signal.NotifyContext`) immediately after creating the context. Inside a loop, put the context and its `defer` in a helper function (see `runCommand`).
- Exempt: purely in-memory code (cache `Add`/`Get`) and blocking terminal input (`lineReader.ReadLine`), which can't be cancelled; Ctrl-C and Ctrl-D end it with `io.EOF`.
- Ctrl-C during a command cancels that command's context and prints `Canceled`; the REPL keeps running. A cancelled request is never cached.
- In tests use `t.Context()`.

## Concurrency
- Document the lifecycle of every goroutine, including what ends it, in a comment.
- Anything that starts a goroutine must offer `Stop` or `Close`. `pokecache.Cache.Stop` and `pokeapi.Client.Close` do; `main` defers `Close`, and tests register it with `t.Cleanup`.
- Types containing a `sync.Mutex` are used through pointers and never copied.

## Code style
- Names: exported `MixedCaps`, unexported `mixedCaps`, initialisms in capitals (`URL`, `ID`). Packages are lowercase single words without underscores.
- Define interfaces in the consuming package and keep them small. Single-method interfaces use the `-er` suffix (`lineReader`); larger ones are named for their role (`API`). Accept interfaces, return concrete types.
- Every package has a package comment, and every exported identifier has a doc comment. Comments explain why, not what.
- No `init()` for global state; use explicit constructors (`New`, `NewClient`, `getCommands`).
- Add a third-party dependency only with a clear justification.

## Testing
- Use only the standard `testing` package. Do not add `testify`. Use `t.Fatalf` when continuing is pointless and `t.Errorf` otherwise; mark helpers with `t.Helper()`.
- Use table-driven tests, with `t.Run` names, for functions with several input cases.
- Always run with `-race`.
- Unit tests never touch the network. Fake HTTP with `pokeapi.WithTransport` and `WithBaseURL`; fake the API in `repl` tests with `fakeAPI`; capture output in a `bytes.Buffer` and set `cfg.roll`.
- Do not use `time.Sleep` to synchronize. Inject a clock (see `pokecache.newCache`). A short bounded sleep is acceptable only to assert that nothing happens.
- Tests that need the real PokéAPI use the `//go:build integration` tag and run with `go test -tags integration ./...`. None exist today; keep them out of the default run.

## Don't
- Discard errors with `_` in production code (outside the exceptions above).
- Store a context in a struct, or start a goroutine without documenting how it ends.
- Call `fmt.Println`, `os.Exit` or `math/rand` directly inside commands.
- Use `init()` for global state.
- Commit unless asked, create branches, or fix unrelated issues. The built binary `pokedexcli` and `repl.log` are gitignored; keep them out of commits.
