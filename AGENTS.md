# opencode instructions

## Coverage

When running `go test` with `-coverprofile`, always use names matching `coverage*.out` or `coverage*.html`:

```bash
go test -coverprofile=coverage.out ./...             # Full run
go test -coverprofile=coverage_lexer.out ./lexer     # Targeted run
go tool cover -html=coverage.out -o coverage.html    # HTML view
```

- Full run: `coverage.out` / `coverage.html`
- Targeted: `coverage_<purpose>.out` / `coverage_<purpose>.html` (e.g. `coverage_lexer.out`)

## Temp Files

Use `.opencode/tmp/` for any temporary files (instead of `/tmp`).

## Style

Only use hyphens (-), never em dashes (—) or en dashes (–).
