# CLI & Configuration Reference

## CLI Flags

All flags support both `--flag value` and `--flag=value` syntax.

### `--errors`

Controls when errors are displayed.

| Value | Description |
|-------|-------------|
| `parse` | Show at parse/analysis time only |
| `run` | Show at runtime only |
| `both` | Both parse and run time (default) |

```
./interp --errors parse program.lang
```

### `--warnings`

Controls when warnings are displayed.

| Value | Description |
|-------|-------------|
| `parse` | Show at parse/analysis time only (default) |
| `run` | Show at runtime only |
| `both` | Both parse and run time |

```
./interp --warnings both program.lang
```

### `--null`

Controls null-operation detection.

| Value | Description |
|-------|-------------|
| `error` | Report as error |
| `warning` | Report as warning (default) |
| `none` | Suppress |

```
./interp --null error program.lang
```

### `--help` / `-h`

Print help text.

## Config File: `lang-config.json`

Read from the current directory at startup. CLI flags override config values.

```json
{
    "errors": "parse",
    "warnings": "both",
    "null": "warning"
}
```

Omitted fields use their defaults.

## Static Analysis

A static analysis pass runs before execution. It detects:

- **Null safety** — using potentially-null values in operations
- **Division by zero** — definite and potential zero divisors
- **Guard analysis** — refines type info from `if` conditions
- **Return type checking** — verifies all paths return a value

```
var x: int{nullable: true} = null;
var y: int = x;  // warning: assigning nullable to non-nullable

var z: int = 0;
var w: int = 10 / z;  // error: division by zero
```

Errors and warnings include source line numbers with source context.
