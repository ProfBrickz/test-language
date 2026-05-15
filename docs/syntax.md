# Syntax

## Declaring Variables

```
var name: type = expression;
```

The initializer is optional. Leave it off and nullable types default to `null`, integers to `0`, floats to `0.0`, and booleans to `false`.

## If / Else

```
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

```
var x: int{size: 32} = 1;

if (true) {
    var y: int{size: 32} = 2;  // scoped to this block
    x = 3;                      // modifies outer x
}

print(x);  // 3
print(y);  // error: y is not defined
```

## Print

```
print(expression);
```

Prints the value to stdout followed by a newline.

## Assignment

```
identifier = expression;
identifier += expression;
identifier -= expression;
identifier *= expression;
identifier /= expression;
```

Operator assignments (`+=`, `-=`, `*=`, `/=`) read the current value, apply the operation, and store the result back.

## Increment / Decrement

```
identifier++;
identifier--;
```

Postfix `++` and `--` increment or decrement an integer or float variable by 1. Prefix forms (`++i`, `--i`) are not supported. Works on both integers and floats. Overflow is checked at runtime.

## For Loops

```
for (init; condition; update) {
    body;
}
```

C-style for loop. All three parts are optional (`for (;;)` is an infinite loop). The init can be a `var` declaration or an assignment. The condition must be a `bool`. The update can be an assignment or `++`/`--`.

The init, condition, and update share a scope. Each iteration creates a new scope for the body. Variables declared in the body are not visible in the condition or update.

## For-In Loops

```
for (var elem in iterable) {
    body;
}
```

Iterates over elements of an array, list, or string. For strings, each element is a single-character string.

## For-At Loops

```
for (var index at iterable) {
    body;
}
```

Iterates over indices (0-based integers) of an array, list, or string.

## For-Of Loops

```
for (var index, value of iterable) {
    body;
}
```

Iterates over both the index and value of an array, list, or string. The first variable receives the index (integer), the second receives the value.

## While Loops

```
while (condition) {
    body;
}
```

Condition must be a `bool`. Each iteration creates a new scope for the body.

## Break / Skip

```
break;
skip;
```

`break` exits the current loop. `skip` skips to the next iteration (like `continue` in C). Both must appear inside a loop. Using them outside a loop is an error.

## Scoping

Variables declared inside any block (`{ }`) are scoped to that block:

- Variables declared inside an `if`, `else`, `for`, or `while` body are not visible outside
- Outer variables are accessible from within blocks
- Assignments to outer variables modify the outer variable (not a shadow)

For for loops, variables declared in the `init` are scoped to the entire loop construct
and are accessible in the condition, update, and body, but not after the loop.

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

## Indexed Assignment

Assign to an element of an array or list by index:

```
arr[0] = 42;
arr[i] = val;
```

Supports compound operators too:

```
arr[0] += 5;
```

## Method Call Statements

Call methods on arrays and lists as standalone statements:

```
a.add(42);
a.add(2, 1);
a.remove(0);
```

The `.add()` method appends a value (or inserts at an index). The `.remove()` method removes an element at the given index and returns it.

## Index Expression

Access an element of an array or list by index in any expression:

```
print(arr[0]);
var x: int = arr[i];
```

## Array / List Literals

```
[1, 2, 3]     // array/list literal with three elements
[]             // empty array/list literal
```

## Member Access

The `.` operator accesses properties of types and values:

```
print(int.min);              // type property: -9223372036854775808
print(float.precision);      // type property: 15
print(float{size: 32}.max);  // type property with parameters: 3.4028235e+38
print(bool.size);            // type property: 8
```

The `typeof` operator returns a value's type descriptor, which can be further accessed:

```
var a: int{size: 8} = 42;
print(typeof(a));       // "8-bit signed int"
print(typeof(a).min);   // -128
print(typeof(a).max);   // 127
```

## Identifiers

Matching `[a-zA-Z_][a-zA-Z0-9_]*`. Used for variable names.
