# References, Copy, Type and Reference Checks

Language builtins for working with references, copying, and type checking at runtime.

---

## `ref` — Create an Alias

Creates an alias to an existing variable. All references to the same storage share mutations.

```
var x: int{size: 32} = 42;
ref r: int{size: 32} = ref x;
print(r);  // 42

r = 100;
print(x);  // 100 (x changed too)
```

### Reference Chaining

Multiple aliases point to the same storage location:

```
var a: int{size: 32} = 10;
ref b: int{size: 32} = ref a;  // b references a's storage
ref c: int{size: 32} = ref b;  // c also references a's storage

a = 100;
print(b);  // 100
print(c);  // 100
```

### Restrictions

- **Cannot assign literals to references**: `ref x: int = 1;` ❌
- **Requires `=` assignment**: The `ref` type must be declared with a variable
- **No indexed assignment**: Works only with variable assignment, not array element assignment
- **Must reference existing variables**: Cannot create references from other literals

---

## `copy` — Shallow Copy

Creates a shallow copy of a value. For arrays, produces an independent copy with separate storage.

```
var a: array{size: 3}<int> = [1, 2, 3];
var b: array{size: 3}<int> = copy a;

b[0] = 99;
print(a[0]);  // 1 (a unchanged)
```

### Behavior by Type

- **Arrays**: Creates a new array with copied elements
- **Other types**: Returns the original value unchanged (copy semantics apply but primitives don't have separate storage)

---

## `isSameStorage` — Reference Identity Check

Tests whether two values reference the same memory storage location.

```
var a: int{size: 32} = 42;
ref r: int{size: 32} = a;
print(r isSameStorage a);  // true (same storage)
print(a isSameStorage r);  // true (same storage)

ref s: int{size: 32} = 100;
print(r isSameStorage s);  // false (different storage)
```

Returns `true` only if both operands reference the same storage location in memory. Useful for detecting when variables are aliases.

---

## `isSameStorage` — Reference Identity Check

Tests whether two variables are aliases (point to the same memory location).

```
var a: int{size: 32} = 42;
ref r: int{size: 32} = ref a;

print(r isSameStorage a);  // true
```

### Reference Chaining

Multiple references to the same storage:

```
var x: int{size: 32} = 100;
ref a: int{size: 32} = ref x;
ref b: int{size: 32} = ref a;

print(a isSameStorage x);  // true
print(b isSameStorage x);  // true
print(a isSameStorage b);  // true
```

### Use Cases

#### Detecting Aliases

```
var original: int{size: 32} = 42;
ref copy: int{size: 32} = ref original;

if (copy isSameStorage original) {
    print("copy is an alias of original");
}
```

#### Avoiding Unnecessary Copies

```
var source: int{size: 32} = 10;
ref dest: int{size: 32} = ref source;

if (!dest isSameStorage source) {
    // Source and dest are different - safe to modify
    dest = 100;
}
```

---

## Next

Learn about [Type Parameters](13-type-parameters.md).
