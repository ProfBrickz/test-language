# Types

Every variable must declare its type at definition. The three primitive types are `int`, `float`, and `bool`, each with optional parameters in curly braces.

## int

```go
var x: int{size: 32, signed: true, nullable: false} = 42;
```

- **size** — how many bits: `8`, `16`, `32`, or `64` (default `64`)
- **signed** — whether it can be negative: `true` or `false` (default `true`)
- **nullable** — whether it can hold `null`: `true` or `false` (default `true`)

## float

```go
var a: float{size: 32, nullable: false} = 3.14;
```

- **size** — `16` (half), `32` (float), or `64` (double) (default `64`)
- **nullable** — `true` or `false` (default `true`)

## bool

```go
var x: bool{nullable: false} = true;
```

- **nullable** — `true` or `false` (default `true`)

## Shorthand

When using all defaults, drop the braces:

```go
var x: int = 42;      // int{size: 64, signed: true, nullable: true}
var a: float = 3.14;  // float{size: 64, nullable: true}
var b: bool = true;    // bool{nullable: true}
```

## Nullability

`nullable: true` means the variable can be `null`. Non-nullable types reject `null` at runtime.

```go
var x: int{nullable: true} = null;   // fine
var y: int{nullable: false} = null;  // runtime error
```

## Automatic Conversion

Values are automatically converted in assignments and operator assignments following common-sense widening rules:

- **Same type** — always fine
- **Integer to integer** — widening works (e.g. `int8` to `int32`); narrowing does not; switching signedness at the same size does not
- **Float to float** — widening works (e.g. `float16` to `float64`); narrowing does not
- **Integer to float** — small int types can widen to floats (`int8`→`float16`, `int16`→`float32`, `int32`→`float64`); `int64` and `uint64` cannot
- **Float to integer** — never automatic
- **Nullable** — you can assign a non-nullable value to a nullable variable, but not the other way around

## Overflow

When a value exceeds its target range, it wraps around:

```go
var x: int{size: 8, signed: true} = 200;   // -56
var y: int{size: 8, signed: false} = 256;  // 0
var a: float{size: 16} = 70000.0;          // +Inf
```

Division by zero is a runtime error for integers and produces infinity for floats.
