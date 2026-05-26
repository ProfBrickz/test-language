# Overview

An expression-based, statically-typed programming language with configurable primitive types.

```
print("Hello, World!");
```

## Quick Tour

- **Configurable types** — integers with explicit size (8/16/32/64), signedness, and nullability; floats with configurable precision; booleans with nullability
- **Strings** — UTF-8 with size constraints, methods, escape sequences
- **Arrays & Lists** — fixed-size and variable-size sequences
- **Functions** — typed parameters, return types, overloading
- **References** — aliases to variables, shallow copy, type checking
- **Unions** — variables that can hold multiple types
- **Structs** — named collections of typed fields with methods
- **Classes** — reference types with inheritance, virtual methods, access control, interfaces, static members
- **No dependencies** beyond the standard library

## How It Works

1. **Lexer** — splits source text into tokens
2. **Parser** — builds an abstract syntax tree
3. **Static Analyzer** — checks for null safety, division by zero, type errors
4. **Interpreter** — walks the tree and runs it directly

## What's Next

Start with [Getting Started](02-getting-started.md) to install and run your first program.

---

*See also: [CLI & Config](cli.md)*
