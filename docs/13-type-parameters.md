# Type Parameters

Type parameters configure the behavior of primitive types inside curly braces.

## Parameter Reference

| Type    | Parameters                          |
|---------|-------------------------------------|
| int     | `size`, `signed`, `nullable`        |
| float   | `size`, `nullable`                  |
| bool    | `nullable`                          |
| string  | `size`, `min`, `max`                |
| array   | `size` (using `{size: N}` or `[N]`) |
| list    | `min`, `max`                        |

## The `auto` Keyword

Use `auto` to mean unbounded or unconstrained:

```
var s: string{size: auto};              // unbounded
var l: list{min: auto, max: auto}<int>; // unbounded
```

When parameters are omitted entirely, they default to `auto`.

## Empty Braces

Empty braces use all defaults:

```
var x: int{} = 42;  // same as int{size: 64, signed: true, nullable: true}
```

## Overflow

Values that exceed their type's range produce a runtime error:

```
var x: int{size: 8, signed: true} = 200;   // error: overflows
var y: int{size: 8, signed: false} = 256;  // error: overflows
var a: float{size: 16} = 70000.0;          // error: overflows
```

Division by zero is a runtime error for integers (produces infinity for floats).

## Type Members

Types expose properties via dot access:

```
print(int.min);                 // -9223372036854775808
print(int{size: 8}.max);        // 127
print(float.precision);         // 15
print(bool.size);               // 8
print(float{size: 32}.min);     // -3.4028235e+38
```

The `typeof` operator returns a type descriptor from a value:

```
var a: int{size: 8} = 42;
print(typeof(a));       // "8-bit signed int"
print(typeof(a).min);   // -128
print(typeof(a).max);   // 127
```

### int Members

| Member | Description | Example | Result |
|--------|-------------|---------|--------|
| `.min` | Minimum value | `int{size:8,signed:true}.min` | `-128` |
| `.max` | Maximum value | `int{size:8,signed:false}.max` | `255` |
| `.size` | Bit width | `int{size:32}.size` | `32` |

### float Members

| Member | Description | Example | Result |
|--------|-------------|---------|--------|
| `.min` | Most negative | `float.min` | `-1.8e+308` |
| `.max` | Largest positive | `float.max` | `1.8e+308` |
| `.min_subnormal` | Smallest positive subnormal | `float.min_subnormal` | `5e-324` |
| `.min_normal` | Smallest positive normal | `float{size:32}.min_normal` | `1.175e-38` |
| `.precision` | Decimal digits | `float{size:16}.precision` | `3` |
| `.min_exponent` | Min exponent | `float{size:32}.min_exponent` | `-126` |
| `.max_exponent` | Max exponent | `float{size:64}.max_exponent` | `1023` |
| `.size` | Bit width | `float.size` | `64` |

### bool Members

| Member | Description | Example | Result |
|--------|-------------|---------|--------|
| `.size` | Bit width | `bool.size` | `8` |

## Next

Learn about [Union Types](14-unions.md).
