# test-language

A vibe coded proof of concept programming language. Everything here is vibe coded — the code, the comments, the commits, the docs, even this sentence.

Inspired by **C**, **Java**, **JavaScript**, **C#**, and **Odin**, with some special features of its own:

- **Configurable primitive types** — Integers with explicit size (8/16/32/64), signedness, and nullability; floats with configurable precision (16/32/64); booleans with nullability
- **Rich literal syntax** — Hex (`0xFF`), binary (`0b1010`), octal (`0o777`), underscore separators (`100_000`), hex/binary/octal floats (`0xf.f`, `0b1.01`, `0o7.7`), scientific notation
- **Well-defined automatic conversion rules** — Widening conversions only, with safe numeric promotion across types
- **No dependencies beyond the standard library** (except `float16`) — small, simple tree-walking interpreter

See [`docs/`](docs/) for full language documentation.

## Quick Start

```bash
go build -o interp .
./interp examples/hello.lang
```
