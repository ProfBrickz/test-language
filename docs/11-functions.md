# Functions

## Declaration

```
function name(param1: type, param2: type): returnType {
    body;
}
```

- Parameters are typed
- Parameters can have default values
- Return type comes after a colon (optional — omit for void functions)
- Body is a block `{ }`

### Examples

```
function add(a: int{size: 32}, b: int{size: 32}): int{size: 32} {
    return a + b;
}

function greet(name: string) {
    print("Hello, " + name + "!");
}
```

## Calling

```
var result: int{size: 32} = add(1, 2);
greet("world");
```

### Default Parameters

Parameters can have default values, which are used when no argument is provided:

```
function greet(name: string = "Guest"): string {
    return "Hello, " + name + "!";
}

function add(a: int, b: int = 0): int {
    return a + b;
}

print(greet());           // Hello, Guest!
print(greet("Alice"));    // Hello, Alice!
print(add(5));            // 5
print(add(5, 3));         // 8
```

Defaults are evaluated at call time. You can use simple expressions, constant variables, or even pure functions with known parameters.

```
const zero: int = 0;

function getNumber(a: int, b: int) {
	return a + (b * 3)
}

function myFunction(a: int = zero, b: int = getNumber(1, 2), message: string = "Hello, " + "world!") {
	print(message);
	return a + b;
}
```

Calls can be used as expressions (returning a value) or as statements (value discarded).

## Return

```
return expression;
return;  // void return
```

- `return expr;` exits the function with a value
- Bare `return;` exits a void function
- Return outside a function is a runtime error
- Functions with a declared return type must return a value on all paths

## Hoisting

Functions are hoisted — callable before declaration:

```
print(double(5));  // 10

function double(x: int): int {
    return x * 2;
}
```

## Overloading

Multiple functions with the same name but different parameter types:

```
function describe(x: int): string {
    return "integer: " + x;
}
function describe(x: string): string {
    return "string: " + x;
}

print(describe(42));     // integer: 42
print(describe("hi"));   // string: hi
```

### Resolution Order

1. **Exact match** — parameter types match exactly
2. **Same-category conversion** — widening within same type (e.g. int8 to int32)
3. **Cross-category conversion** — widening across types (e.g. int8 to float32)

If no overload matches, it is a compile-time error.

### Ambiguous Null Calls

When multiple overloads could accept `null`, direct use of a `null` literal is ambiguous and rejected. However, you can pass a variable set to `null` or use type conversion on a null literal:

```
function foo(x: int{nullable: true}): void { }
function foo(x: string{nullable: true}): void { }

var n: int{nullable: true} = null;
foo(n);  // OK: n is explicitly typed as nullable int

foo((int{nullable: true})null);  // OK: type conversion disambiguates
foo(null);  // error: ambiguous null literal
```

The ambiguity arises because the compiler cannot determine which nullable type to use from a bare `null` literal when multiple overloads are possible.

## Next

Learn about [References, Copy, and Is](12-references.md).
