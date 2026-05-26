# References, Copy, and Is

## ref (Reference)

Creates an alias to an existing variable. Mutations through any alias affect the same storage.

```
var x: int{size: 32} = 42;
ref r: int{size: 32} = x;
print(r);  // 42

r = 100;
print(x);  // 100 (x changed too)
```

### Reference Assignment

Redirect a variable to point to another variable's storage:

```
var a: int{size: 32} = 1;
var b: int{size: 32} = 2;

a = ref b;  // a now points to b's storage
a = 100;
print(b);  // 100
```

Only works with `=`. Cannot be used with indexed assignment.

## copy (Shallow Copy)

Creates a shallow copy of a value. For arrays, produces an independent copy.

```
var a: array{size: 3}<int> = [1, 2, 3];
var b: array{size: 3}<int> = copy a;

b[0] = 99;
print(a[0]);  // 1 (independent)
```

For non-array types, returns the value unchanged.

## is (Reference Identity)

```
var a: int{size: 32} = 42;
ref r: int{size: 32} = a;
print(a is r);  // true (same storage)
```

Returns `true` only if both sides refer to the same storage in memory.

## instanceOf (Type Check)

```
var val: int | string = "hello";
print(val instanceOf string);  // true
print(val instanceOf int);     // false
```

Checks the runtime type against a type reference. Works with all types — primitives, structs, and classes. For class types, `instanceOf` is inheritance-aware (returns `true` if the value is an instance of the type or any of its subclasses).

## Next

Learn about [Type Parameters](13-type-parameters.md).
