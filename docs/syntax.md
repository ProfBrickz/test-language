# Syntax

## Declaring Variables

```go
var name: type = expression;
```

The initializer is optional. Leave it off and nullable types default to `null`, integers to `0`, floats to `0.0`, and booleans to `false`.

## Print

```go
print(expression);
```

Prints the value to stdout followed by a newline.

## Assignment

```go
identifier = expression;
identifier += expression;
identifier -= expression;
identifier *= expression;
identifier /= expression;
```

Operator assignments (`+=`, `-=`, `*=`, `/=`) read the current value, apply the operation, and store the result back.

## Operators

Arithmetic: `+`, `-`, `*`, `/`. Multiplication and division bind tighter than addition and subtraction. Use parentheses `( )` to override.

## Literals

**Integers:** plain decimal (`42`), hex (`0xFF`), binary (`0b1010`), octal (`0o777`), and underscore separators (`100_000`).

**Floats:** decimal (`3.14`), scientific (`1.5e3`), hex float (`0xf.f`), binary float (`0b1.01`), octal float (`0o7.7`), and underscores (`1_000.5`).

**Booleans:** `true` and `false`.

**Null:** `null`.

## Identifiers

Matching `[a-zA-Z_][a-zA-Z0-9_]*`. Used for variable names.
