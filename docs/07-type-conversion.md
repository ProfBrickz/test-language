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

// Integer to float
var c: int{size: 8} = 10;
var d: float{size: 32} = c;             // OK: int8 to float32

// Float widening
var e: float{size: 16} = 1.5;
var f: float{size: 32} = e;             // OK: float16 to float32

// Non-nullable to nullable
var g: int{nullable: false} = 5;
var h: int{nullable: true} = g;         // OK
```

## Next

Learn about [Arrays and Lists](08-arrays-and-lists.md).
