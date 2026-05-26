# Getting Started

## Build

```bash
go build -o interp .
```

## Run a File

```bash
./interp program.lang
```

## REPL

Start the interactive REPL (no filename):

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

## Hello World

Create `hello.lang`:

```
print("Hello, World!");
```

Run it:

```bash
./interp hello.lang
```

## CLI Help

```
./interp --help
```

## Next

Learn [Basic Syntax](03-syntax.md).
