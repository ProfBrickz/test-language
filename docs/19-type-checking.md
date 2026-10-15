# Type Checking

Runtime type checking for values, determining their actual type at runtime.

---

## `isType` — Type Compatibility Check

Tests whether a value's runtime type matches (or is compatible with) a given type.

```
var val: int | string = "hello";
print(val isType string);  // true
print(val isType int);     // false
```

### Union Types

Type checking works naturally with union types:

```
var data: int | string | bool = 42;
print(data isType int);        // true
print(data isType string);     // false
print(data isType bool);       // false
```

The union form is equivalent to checking each member individually:

```
// Equivalent forms:
let x: int | string = 42;
print((x isType string) || (x isType int));  // true (via two checks)
print(x isType string | int);                 // true (via union type)
```

### Subtype Compatibility

Works with subtype relationships:

```
type Animal = { name: string };
type Dog = { name: string; bark: () -> void };

var a: Animal = { name: "Fluffy" };
print(a isType Animal);  // true
print(a isType Dog);     // false (missing bark method)
```

### Practical Use Cases

#### Conditional Processing

```
var input: int | string = "123";

if (input isType int) {
    print("Got an integer");
} else if (input isType string) {
    print("Got a string");
}
```

#### Pattern Matching

```
var val: int | string;

if (val isType int) {
    let n = val as int;
    print(n * 2);
}
```

#### Runtime Type Guards

```
function process(value: int | string): void {
    if (value isType int) {
        // Handle integer case
        print("Processing number");
    } else {
        // Handle string case
        print("Processing text");
    }
}
```

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

## Comparison

| Operation       | Checks                     | Example             |
| --------------- | -------------------------- | ------------------- |
| `isType`        | Runtime type compatibility | `val is int`        |
| `isSameStorage` | Reference identity (alias) | `a isSameStorage b` |

- **`isType`** answers: "What type is this value?"
- **`isSameStorage`** answers: "Do these variables share memory?"
