# Variables

## Declaration

Every variable must declare its type:

```
var name: type = expression;
```

The initializer is optional. Defaults:

| Type    | Default        |
|---------|----------------|
| int     | `0` (nullable) or `null` |
| float   | `0.0` (nullable) or `null` |
| bool    | `false` (nullable) or `null` |
| string  | `""`           |

## Basic Examples

```
var x: int = 42;
var y: float = 3.14;
var flag: bool = true;
var name: string = "hello";
```

## Primitive Types

### int

```
var x: int = 42;        // int{size: 64, signed: true, nullable: true}
var y: int{size: 32} = 42;
var z: int{nullable: false} = 0;
```

Parameters:
- **size** — bits: `8`, `16`, `32`, or `64` (default `64`)
- **signed** — `true` or `false` (default `true`)
- **nullable** — `true` or `false` (default `true`)

### float

```
var a: float = 3.14;        // float{size: 64, nullable: true}
var b: float{size: 32} = 3.14;
```

Parameters:
- **size** — `16` (half), `32` (float), or `64` (double) (default `64`)
- **nullable** — `true` or `false` (default `true`)

### bool

```
var x: bool = true;         // bool{nullable: true}
var y: bool{nullable: false} = false;
```

Parameters:
- **nullable** — `true` or `false` (default `true`)

### Shorthand

Drop the braces when using all defaults:

```
var x: int = 42;      // int{size: 64, signed: true, nullable: true}
var a: float = 3.14;  // float{size: 64, nullable: true}
var b: bool = true;    // bool{nullable: true}
```

### Nullable Types

```
var x: int{nullable: true} = null;
var y: int{nullable: true};      // defaults to null
var z: int{nullable: false} = 0; // defaults to 0
```

## Assignment

```
x = 20;
x += 5;   // x = x + 5
x -= 3;   // x = x - 3
x *= 2;   // x = x * 2
x /= 4;   // x = x / 4
x %= 3;   // x = x % 3
```

## Increment / Decrement

```
var i: int = 5;
i++;
print(i);  // 6
i--;
print(i);  // 5
```

Postfix only (prefix `++i` is not supported). Works on ints and floats.

## Next

Learn about [Operators](05-operators.md).
