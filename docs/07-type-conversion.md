# Type Conversion

Values are automatically converted in assignments and operator assignments following widening rules.

## Rules

- **Same type** — always fine
- **Integer to integer** — widening works (e.g. int8 to int32); narrowing does not; switching signedness at the same size does not
- **Float to float** — widening works (e.g. float16 to float64); narrowing does not
- **Integer to float** — small int types can widen to floats (int8 to float16, int16 to float32, int32 to float64); int64 and uint64 cannot
- **Float to integer** — never automatic
- **Integer/float to bool** — never automatic
- **Nullable** — non-nullable can be assigned to nullable, but not the reverse

## Examples
```
// Integer widening
var a: int{size: 8, signed: true} = 10;
var b: int{size: 32, signed: true} = a;  // OK: int8 to int32

// Explicit type conversion (C-style syntax)
var c: int{size: 8} = 10;
var d: float{size: 32} = (float{size: 32})c;  // OK: explicit conversion
var e = (int)c;  // implicit narrowing not allowed

// Float widening
var e: float{size: 16} = 1.5;
var f: float{size: 32} = e;             // OK: float16 to float32

// Non-nullable to nullable
var e: float{size: 16} = 1.5;
var f: float{size: 32} = e;             // OK: float16 to float32

// Non-nullable to nullable
var g: int{nullable: false} = 5;
var h: int{nullable: true} = g;         // OK
```

## Optional Chaining and Null-Coalescing

### Optional Chaining (`?.`)

Use `?.` to safely access properties or call methods on potentially null values. Returns `null` if the chain is broken:

```
// Safe property access
var name = user?.name;           // null if user is null
var email = user?.profile?.email; // null if profile is null

// Safe method calls
var count = user?.getName?.();   // null if getName is null or user is null

// Optional member access in expressions
var displayName = user?.displayName ?? "Anonymous";
```

### Null-Coalescing (`??`)

Use `??` to provide a default value when a value is `null`:

```
// Basic null-coalescing
var name = userInput ?? "default";          // "default" if userInput is null
var age = ageInput ?? 0;                    // 0 if ageInput is null
var config = configValue ?? {timeout: 30};  // default object if null
```

The right-hand side is evaluated only when the left-hand side is `null`.

### Combined Usage

```
// Common pattern: safe access with default
var displayName = user?.displayName ?? "Guest";
var email = user?.email ?? "user@example.com";
var status = user?.getStatus?.() ?? "unknown";
```

## Next

Learn about [Arrays and Lists](08-arrays-and-lists.md).
