# Types

Every variable must declare its type at definition. The three primitive types are `int`, `float`, and `bool`, each with optional parameters in curly braces.

## int

```
var x: int{size: 32, signed: true, nullable: false} = 42;
```

- **size** - how many bits: `8`, `16`, `32`, or `64` (default `64`)
- **signed** - whether it can be negative: `true` or `false` (default `true`)
- **nullable** - whether it can hold `null`: `true` or `false` (default `true`)

## float

```
var a: float{size: 32, nullable: false} = 3.14;
```

- **size** - `16` (half), `32` (float), or `64` (double) (default `64`)
- **nullable** - `true` or `false` (default `true`)

## bool

```
var x: bool{nullable: false} = true;
```

- **nullable** - `true` or `false` (default `true`)

## Shorthand

When using all defaults, drop the braces:

```
var x: int = 42;      // int{size: 64, signed: true, nullable: true}
var a: float = 3.14;  // float{size: 64, nullable: true}
var b: bool = true;    // bool{nullable: true}
```

## Nullability

`nullable: true` means the variable can be `null`. Non-nullable types reject `null` at runtime.

```
var x: int{nullable: true} = null;   // fine
var y: int{nullable: false} = null;  // runtime error
```

**Null in comparisons:**

- `null == null` → `true`
- `null != null` → `false`
- `null == value`, `value == null` → `false`
- `null != value`, `value != null` → `true`
- `null < value`, `null > value`, etc. → `false` (warning when using literal `null`)

**Null in boolean operations:**

- `!null` → `true` (null is falsy, warning when using literal `null`)
- `null && value` → `false` (null is falsy, warning when using literal `null`)
- `value && null` → `false` (short-circuits if value is false, warning when using literal `null`)
- `null || value` → `value` (null is falsy, warning when using literal `null`)
- `value || null` → `true` (short-circuits if value is true, warning when using literal `null`)

## Automatic Conversion

Values are automatically converted in assignments and operator assignments following common-sense widening rules:

- **Same type** - always fine
- **Integer to integer** - widening works (e.g. `int8` to `int32`); narrowing does not; switching signedness at the same size does not
- **Float to float** - widening works (e.g. `float16` to `float64`); narrowing does not
- **Integer to float** - small int types can widen to floats (`int8`→`float16`, `int16`→`float32`, `int32`→`float64`); `int64` and `uint64` cannot
- **Float to integer** - never automatic
- **Integer/float to bool** - never automatic
- **Nullable** - you can assign a non-nullable value to a nullable variable, but not the other way around

## Operator Type Restrictions

- **Arithmetic (`+`, `-`, `*`, `/`)**: integers and floats only
- **Equality (`==`, `!=`)**: integers, floats, booleans, and null
- **Ordering (`<`, `>`, `<=`, `>=`)**: integers and floats only (not booleans)
- **Logical (`&&`, `||`, `!`)**: booleans only

## Overflow

When a value exceeds its target range, a runtime error is raised:

```
var x: int{size: 8, signed: true} = 200;   // error: overflows 8-bit signed int
var y: int{size: 8, signed: false} = 256;  // error: overflows 8-bit unsigned int
var a: float{size: 16} = 70000.0;          // error: overflows 16-bit float
```

Division by zero is a runtime error for integers and produces infinity for floats.

## Type Members

Types expose properties via dot access. The default parameters are used when no braces are specified.

### int members

| Member   | Description            | Example                          | Result                     |
|----------|------------------------|----------------------------------|----------------------------|
| `.min`   | Minimum representable  | `int{size: 8, signed: true}.min` | `-128`                     |
| `.max`   | Maximum representable  | `int{size: 8, signed: false}.max`| `255`                      |
| `.size`  | Bit width              | `int{size: 32}.size`             | `32`                       |

Default `int` is `int{size: 64, signed: true}`.

### float members

| Member            | Description                    | Example                              | Result          |
|-------------------|--------------------------------|--------------------------------------|-----------------|
| `.min`            | Minimum (most negative)        | `float.min`                          | `-1.7976931348623157e+308` |
| `.max`            | Maximum (largest positive)     | `float.max`                          | `1.7976931348623157e+308` |
| `.min_subnormal`  | Smallest positive subnormal   | `float.min_subnormal`                | `5e-324`         |
| `.min_normal`     | Smallest positive normal       | `float{size: 32}.min_normal`         | `1.1754944e-38`  |
| `.precision`      | Decimal digits of precision    | `float{size: 16}.precision`          | `3`              |
| `.min_exponent`   | Minimum exponent               | `float{size: 32}.min_exponent`       | `-126`           |
| `.max_exponent`   | Maximum exponent               | `float{size: 64}.max_exponent`       | `1023`           |
| `.size`           | Bit width                      | `float.size`                         | `64`             |

Default `float` is `float{size: 64}`.

### bool members

| Member   | Description | Example        | Result |
|----------|-------------|----------------|--------|
| `.size`  | Bit width   | `bool.size`    | `8`    |

### .type member

Every value has a `.type` member that returns a type descriptor, which supports the same members above:

```
var a: int{size: 8} = 42;
print(a.type);       // "8-bit signed int"
print(a.type.min);   // -128
print(a.type.max);   // 127
```

## array

Fixed-size sequence of elements of the same type.

```
var a: array{size: 5}<int> = [1, 2, 3, 4, 5];
```

- **size** - number of elements (required)
- **Element type** - any type, specified in angle brackets `< >`
- **Default initializer**: zero-filled (all elements are `0`)

```
var a: array{size: 3}<int>;
print(a);  // [0, 0, 0]
```

### Array Members

| Member   | Description         | Example                   | Result |
|----------|---------------------|---------------------------|--------|
| `.length`| Number of elements  | `a.length`                | `5`    |
| `.type`  | Type descriptor     | `a.type`                  | `"array{size: 5}<64-bit signed int>"` |

### Array Type Members

| Member   | Description         | Example                              | Result |
|----------|---------------------|--------------------------------------|--------|
| `.size`  | Array capacity      | `array{size: 5}<int>.size`           | `5`    |
| `.length`| Array capacity      | `array{size: 5}<int>.length`         | `5`    |
| `.elem_type` | Element type descriptor | `array{size: 5}<int>.elem_type`  | `"64-bit signed int"` |

Chaining works: `array{size: 5}<int>.elem_type.size` returns `64`.

### Array Shorthand Syntax

Square brackets can be used instead of `{size: N}`:

```
var a: array[5]<int>;      // same as array{size: 5}<int>
```

### Size Inference

When the size is omitted, it is inferred from the initializer:

```
var a: array<int> = [1, 2, 3];  // size inferred as 3
print(a.length);                 // 3
```

## list

Variable-size sequence of elements of the same type.

```
var a: list<int> = [1, 2, 3];
var b: list{min: 1, max: 10}<int> = [1, 2, 3];
```

- **min** - minimum number of elements (optional, default `auto` i.e. unbounded)
- **max** - maximum number of elements (optional, default `auto` i.e. unbounded)
- **Initializer is required when `min > 0`**

### List Members

| Member   | Description                            | Example           | Result        |
|----------|----------------------------------------|-------------------|---------------|
| `.length`| Number of elements                     | `a.length`        | `3`           |
| `.add(v)`| Append value                           | `a.add(42)`       | `null`        |
| `.add(v, i)`| Insert value at index               | `a.add(2, 1)`     | `null`        |
| `.remove(i)`| Remove element at index and return it| `a.remove(0)`     | `1`           |
| `.type`  | Type descriptor                        | `a.type`          | `"list<64-bit signed int>"` |

Note: `.add()` and `.remove()` are statements (called for their side effects). The bounds (`min`/`max`) are enforced at runtime: adding beyond `max` or removing below `min` produces an error.
