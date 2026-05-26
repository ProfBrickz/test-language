# Functions

## Declaration

```
function name(param1: type, param2: type): returnType {
    body;
}
```

- Parameters are typed
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

Passing `null` when multiple overloads could accept it is rejected:

```
function foo(x: int{nullable: true}): void { }
function foo(x: string{nullable: true}): void { }

foo(null);  // error: ambiguous
```

## Next

Learn about [References, Copy, and Is](12-references.md).
