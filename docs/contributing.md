# Contributing

Keep contributions small and focused.

## Run Locally

Make sure `gopass` is installed and already configured on your machine.

```bash
go run .
```

Before sending changes, run:

```bash
gofmt -w .
go test ./...
go build ./...
```

## Project Rules

- Wrap `gopass`; do not reimplement password-store logic in Go.
- Keep the `gopass` command layer separate from the Bubble Tea UI.
- Prefer small, idiomatic Go changes over extra abstraction.
- Return errors instead of hiding them; user-facing failures should keep `gopass` context.
- Add doc comments for exported symbols.

## Commits

- Keep commits atomic.
- Use Conventional Commits when possible, for example `fix(ui): keep preview in sync`.

That is enough for this project: clear changes, simple code, and no extra process.
