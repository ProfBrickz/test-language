# Syntax

## Statements

Every statement ends with a semicolon:

```
print("hello");
var x: int = 1;
x = x + 1;
```

## Comments

Single-line comments use `//`:

```
var x: int = 42; // anything after // is ignored
```

## Blocks

Curly braces group statements into blocks:

```
{
    var x: int = 1;
    print(x);
}
```

Blocks create new scopes (see [Control Flow](09-control-flow.md)).

## Identifiers

Must match `[a-zA-Z_][a-zA-Z0-9_]*`. Used for variable and function names.

```
x
myVar
_count
value2
```

## Next

Learn about [Variables](04-variables.md).
