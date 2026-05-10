# Overview

An expression-based, statically-typed programming language with integers, floats, and booleans that let you configure their size, signedness, and whether they can be null.

## How It Works

Source code goes through three phases:

1. **Lexer** - splits text into tokens
2. **Parser** - builds an abstract syntax tree
3. **Interpreter** - walks the tree and runs it

No bytecode, no VM, no compiler - the tree runs directly.

## Usage

Run a file:

```bash
./interp program.lang
```

Start the REPL:

```bash
./interp
```

```
Welcome to the language interpreter (type 'exit' to quit)
> var x: int{size: 32} = 42;
> print(x);
42
> exit
```

Build from source:

```bash
go build -o interp .
```

## Comments

Single-line comments use `//`:

```
var x: int = 42; // anything after // is ignored
```
