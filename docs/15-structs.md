# Structs

Structs are named collections of typed fields. They are defined at the top level and used as types for variables, parameters, and return values.

## Declaration

```
struct Name {
    field1: type1;
    field2: type2;
    field3: type3;
}
```

- Fields are separated by semicolons
- The trailing semicolon is optional
- Structs are value types — assigned and passed by copy
- Struct names are usable as types after declaration

### Examples

```
struct Point {
    x: int;
    y: int;
}

struct Rectangle {
    top_left: Point;
    bottom_right: Point;
}
```

## Literals

A struct literal creates a value of a named struct type:

```
var p: Point = Point { x: 10, y: 20 };
```

Fields are specified by name. Order does not matter:

```
var a: Point = Point { y: 20, x: 10 };  // same as above
```

### Inferred Literals

When the target type is known from the variable declaration, the type name can be omitted:

```
var p: Point;
p = { x: 1, y: 2 };
```

All required fields must be present. Extra fields are an error.

## Field Access

Use the `.` operator to read or write fields:

```
var p: Point = Point { x: 3, y: 4 };
print(p.x);        // 3
p.y = 10;
```

Nested field access works naturally:

```
var r: Rectangle = Rectangle {
    top_left: Point { x: 0, y: 0 },
    bottom_right: Point { x: 100, y: 100 }
};
print(r.top_left.x);  // 0
```

## Copy with Modifications

The `...` spread syntax creates a new struct from an existing one, overriding select fields:

```
var p1: Point = Point { x: 1, y: 2 };
var p2: Point = { ...p1, y: 5 };
print(p2);  // Point { x: 1, y: 5 }
```

The source struct is unchanged — all its fields are copied, then explicitly listed fields override. This works in any context where a struct literal is expected:

```
var points: Point[] = [
    { ...origin, x: 10 },
    { ...origin, x: 20, y: 30 }
];
```

## Methods

Functions can be defined inside a struct body. The first parameter is the receiver, conventionally named `self`:

```
struct Point {
    x: int;
    y: int;

    function toString(self: Point): string {
        return "(" + self.x + ", " + self.y + ")";
    }

    function magnitude(self: Point): float {
        return sqrt(self.x * self.x + self.y * self.y);
    }
}

var p: Point = Point { x: 3, y: 4 };
print(p.toString());     // (3, 4)
print(p.magnitude());    // 5
```

### Method Calls

- `p.method(args...)` is equivalent to `Point.method(p, args...)`
- The receiver is automatically prepended to the argument list
- Methods follow the same overloading rules as regular functions

## toString

Every struct has a default string representation:

```
var p: Point = Point { x: 10, y: 20 };
print(p);  // Point { x: 10, y: 20 }
```

Define a custom `toString` method to override it:

```
struct Point {
    x: int;
    y: int;

    function toString(self: Point): string {
        return "(" + self.x + ", " + self.y + ")";
    }
}

print(p);  // (10, 20)
```

## Value Semantics

Structs are copied on assignment and when passed to functions:

```
var a: Point = Point { x: 1, y: 2 };
var b: Point = a;
b.x = 99;

print(a.x);  // 1  (unchanged)
print(b.x);  // 99
```

Use `ref` for reference-like behavior:

```
var a: Point = Point { x: 1, y: 2 };
ref b: Point = &a;
b.x = 99;

print(a.x);  // 99
```

## Equality

Structs are compared with `==` and `!=`. Two structs are equal if all their fields are equal. This behavior can be overridden (see [Operator Overloading](18-operator-overloading.md)):

```
var a: Point = Point { x: 1, y: 2 };
var b: Point = Point { x: 1, y: 2 };
print(a == b);  // true
```

## Type Checking with `is`

```
var p: Point = Point { x: 1, y: 2 };
print(p is Point);   // true
print(p is int);     // false
```

## In Functions

Structs can be used as parameter and return types:

```
function midpoint(a: Point, b: Point): Point {
    return Point {
        x: (a.x + b.x) / 2,
        y: (a.y + b.y) / 2
    };
}
```

## Restrictions

- Struct types must be declared before use (no forward references)
- Self-referencing structs are not supported
- Methods cannot mutate the receiver through `self` (it is a copy)
- Anonymous struct types are not supported — all structs must be named

## See Also

- [Operator Overloading](18-operator-overloading.md) — custom operator behavior for structs and classes
- [Properties, Getters, and Setters](17-properties.md) — custom accessor logic for fields
- [Classes](16-classes.md) — reference type alternative to structs
