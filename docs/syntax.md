# Syntax

## Declaring Variables

```go
var name: type = expression;
```

The initializer is optional. Leave it off and nullable types default to `null`, integers to `0`, floats to `0.0`, and booleans to `false`.

## If / Else

```go
if (condition) {
    statements;
} else if (condition) {
    statements;
} else {
    statements;
}
```

Condition must be a `bool`. The `else` and `else if` clauses are optional. Blocks (`{ }`) are required.

Each `if`, `else if`, and `else` body creates a new scope:

- Variables declared inside are not visible outside
- Outer variables are accessible from within
- Assignments to outer variables modify the outer variable (not a shadow)

```go
var x: int{size: 32} = 1;

if (true) {
    var y: int{size: 32} = 2;  // scoped to this block
    x = 3;                      // modifies outer x
}

print(x);  // 3
print(y);  // error: y is not defined
```

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

**Arithmetic:** `+`, `-`, `*`, `/`. Multiplication and division bind tighter than addition and subtraction. Use parentheses `( )` to override.

**Comparison:** `==`, `!=`, `<`, `>`, `<=`, `>=`.

- `==` and `!=` work on integers, floats, booleans, and null
- `<`, `>`, `<=`, `>=` work on integers and floats only (not booleans)

**Logical:** `&&`, `||`, `!`.

- `&&` (AND), `||` (OR), `!` (NOT) work on booleans only
- `&&` and `||` short-circuit: `false && x` returns false without evaluating `x`; `true || x` returns true without evaluating `x`

**Precedence** (highest to lowest):
1. `!` (unary)
2. `*`, `/`
3. `+`, `-`
4. `<`, `>`, `<=`, `>=`
5. `==`, `!=`
6. `&&`
7. `||`

## Literals

**Integers:** plain decimal (`42`), hex (`0xFF`), binary (`0b1010`), octal (`0o777`), and underscore separators (`100_000`).

**Floats:** decimal (`3.14`, `.1`), scientific (`1.5e3`, `.1e2`), hex float (`0xf.f`), binary float (`0b1.01`), octal float (`0o7.7`), and underscores (`1_000.5`, `.5_000`).

Special float literals: `NaN`, `infinity`, `-infinity`.

**Booleans:** `true` and `false`.

**Null:** `null`.

## Member Access

The `.` operator accesses properties of types and values:

```go
print(int.min);              // type property: -9223372036854775808
print(float.precision);      // type property: 15
print(float{size: 32}.max);  // type property with parameters: 3.4028235e+38
print(bool.size);            // type property: 8
```

The `.type` property returns a value's type descriptor, which can be further accessed:

```go
var a: int{size: 8} = 42;
print(a.type);       // "8-bit signed int"
print(a.type.min);   // -128
print(a.type.max);   // 127
```

## Identifiers

Matching `[a-zA-Z_][a-zA-Z0-9_]*`. Used for variable names.
