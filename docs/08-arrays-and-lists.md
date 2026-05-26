# Arrays and Lists

## Arrays

Fixed-size sequence of elements of the same type.

```
var a: array{size: 5}<int> = [1, 2, 3, 4, 5];
```

- **size** — number of elements (required)
- **Element type** — in angle brackets `<>`
- **Default initializer**: zero-filled

### Shorthand

Square brackets instead of `{size: N}`:

```
var a: array[5]<int>;  // same as array{size: 5}<int>
```

### Size Inference

Omit the size and it is inferred from the initializer:

```
var a: array<int> = [1, 2, 3];  // size inferred as 3
print(a.length);                 // 3
```

### Array Members

| Member | Description | Example | Result |
|--------|-------------|---------|--------|
| `.length` | Number of elements | `a.length` | `5` |

### Array Type Members

| Member | Description | Example | Result |
|--------|-------------|---------|--------|
| `.size` | Capacity | `array{size:5}<int>.size` | `5` |
| `.elem_type` | Element type descriptor | `array{size:5}<int>.elem_type` | `"64-bit signed int"` |

Chaining: `array{size:5}<int>.elem_type.size` returns `64`.

## Lists

Variable-size sequence of elements of the same type.

```
var a: list<int> = [1, 2, 3];
var b: list{min: 1, max: 10}<int> = [1, 2, 3];
```

- **min** — minimum elements (optional, default unbounded)
- **max** — maximum elements (optional, default unbounded)
- **Initializer required when `min > 0`**

### List Members

| Member | Description | Example | Result |
|--------|-------------|---------|--------|
| `.length` | Number of elements | `a.length` | `3` |
| `.add(v)` | Append value | `a.add(42)` | `null` |
| `.add(v, i)` | Insert at index | `a.add(2, 1)` | `null` |
| `.remove(i)` | Remove at index | `a.remove(0)` | `1` |

`.add()` and `.remove()` are used as statements. Bounds are enforced at runtime.

## Indexing

```
print(arr[0]);
var x: int = arr[i];
arr[0] = 42;
arr[0] += 5;
```

## Array / List Literals

```
[1, 2, 3]     // array/list literal
[]             // empty literal
```

## Next

Learn about [Control Flow](09-control-flow.md).
