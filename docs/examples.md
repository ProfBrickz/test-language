# Examples

## Hello World

```go
print("Hello, World!");
```

## Variables and Types

```go
// Basic variable declarations
var x: int{size: 32, signed: true} = 42;
var y: float{size: 64} = 3.14;
var flag: bool{nullable: false} = true;

// Using default type parameters
var a: int = 100;       // int{size: 64, signed: true, nullable: true}
var b: float = 2.5;     // float{size: 64, nullable: true}
var c: bool = false;    // bool{nullable: true}
```

## Nullable Types

```go
var x: int{nullable: true} = null;
var y: int{nullable: true};  // defaults to null

var z: int{nullable: false} = 0;  // defaults to 0
```

## Arithmetic

```go
var x: int = 10;
var y: int = 3;

print(x + y);  // 13
print(x - y);  // 7
print(x * y);  // 30
print(x / y);  // 3 (integer division)

var a: float = 10.0;
var b: float = 3.0;

print(a / b);  // 3.3333333333333335
```

## Comparison Operators

```go
// Integer comparisons
print(5 == 5);   // true
print(5 != 3);   // true
print(3 < 5);    // true
print(5 > 3);    // true
print(3 <= 3);   // true
print(5 >= 3);   // true

// Float comparisons
print(3.14 > 2.0);  // true
print(1.5 < 2.5);   // true

// Boolean equality
print(true == true);    // true
print(true != false);   // true

// Null comparisons
print(null == null);   // true
print(null != null);   // false
print(null == 5);      // false
print(null != 5);      // true
print(null < 10);      // false (warning: null literal used with <)
```

## Logical Operators

```go
print(true && true);    // true
print(true && false);   // false
print(false && true);   // false (short-circuit: false && anything is false)

print(true || true);    // true
print(true || false);   // true (short-circuit: true || anything is true)
print(false || true);   // true

print(!true);    // false
print(!false);   // true

// Null with logical operators (null is falsy)
print(!null);              // true (warning: null literal used with !)
print(null && true);       // false (warning: null literal used with &&)
print(false || null);      // false (warning: null literal used with ||)
print(true || null);       // true (short-circuits, right not evaluated)
```

## Operator Precedence

```go
print(1 + 2 * 3);      // 7
print((1 + 2) * 3);    // 9
print(10 - 4 / 2);     // 8
print((10 - 4) / 2);   // 3

// Logical precedence: && before ||
print(true || false && false);  // true (&& binds tighter)
print((true || false) && false);  // false (with parentheses)
```

## Assignments

```go
var x: int{size: 32} = 10;

x = 20;     // simple assignment
x += 5;     // operator: x = x + 5 → 25
x -= 3;     // operator: x = x - 3 → 22
x *= 2;     // operator: x = x * 2 → 44
x /= 4;     // operator: x = x / 4 → 11
```

## Automatic Type Conversion

```go
// Integer widening
var a: int{size: 8, signed: true} = 10;
var b: int{size: 32, signed: true} = a;  // OK: int8 → int32

// Integer to float (within range)
var c: int{size: 8} = 10;
var d: float{size: 32} = c;              // OK: int8 → float32

// Float widening
var e: float{size: 16} = 1.5;
var f: float{size: 32} = e;              // OK: float16 → float32

// Non-nullable to nullable
var g: int{nullable: false} = 5;
var h: int{nullable: true} = g;          // OK
```

## Literal Formats

```go
// Integer literals
var dec: int = 100_000;     // underscores for readability
var hex: int = 0xFF;        // hexadecimal: 255
var bin: int = 0b1010;      // binary: 10
var oct: int = 0o777;       // octal: 511

// Float literals
var sci: float = 1.5e3;            // scientific: 1500.0
var dl: float = .1;                // dot-leading: 0.1
var dl2: float = .1e2;             // dot-leading with exponent: 10.0
var hf: float = 0xf.f;             // hex float: 15.9375
var bf: float = 0b1.01;            // binary float: 1.25
var of: float = 0o7.7;             // octal float: 7.875
var us: float = 1_000.5e1_0;       // underscores

// Special float literals
var nan: float = NaN;               // Not a Number
var inf: float = infinity;           // Positive infinity
var ninf: float = -infinity;         // Negative infinity
```

## Overflow

Overflow is a runtime error:

```go
var x: int{size: 8, signed: true} = 200;   // error: value overflows type
var y: int{size: 8, signed: false} = 256;  // error: value overflows type
```

## Complete Program

```go
// Variable declarations
var x: int{size: 32} = 42;
var y: int{size: 32} = 10;
var result: int{size: 32};

// Arithmetic
result = x + y;
print(result);  // 52

result = x - y;
print(result);  // 32

result = x * y;
print(result);  // 420

result = x / y;
print(result);  // 4

// Operator assignments
result = 100;
result += 50;
print(result);  // 150

// Float example
var a: float{size: 64} = 22.0;
var b: float{size: 64} = 7.0;
print(a / b);  // 3.142857142857143

// Booleans
var flag: bool{nullable: false} = true;
print(flag);   // true

// Nullable
var maybe: int{nullable: true};
print(maybe);  // null

// Comparisons
print(5 == 5);        // true
print(true && false); // false
print(!null);         // true
```

## Type Members

```go
// Integer type properties
print(int.min);                 // -9223372036854775808
print(int{size: 8}.max);        // 127
print(int{size: 8}.size);       // 8

// Float type properties
print(float.precision);         // 15
print(float{size: 32}.min);     // -3.4028235e+38
print(float{size: 16}.max);     // 65504

// Bool type properties
print(bool.size);               // 8

// .type on values
var a: int{size: 8} = 100;
print(a.type);                  // 8-bit signed int
print(a.type.min);              // -128
print(a.type.max);              // 127
```
