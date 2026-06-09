# Operator Overloading

Types can define custom behavior for operators using the `operator` keyword. Structs support the full set of overloadable operators; classes support only `==` and `!=`.

## Syntax

Operators are declared inside a type body using `operator` followed by the operator symbol:

```
operator +(left: Vector2, right: Vector2): Vector2 {
    return Vector2 { x: left.x + right.x, y: left.y + right.y };
}

operator ==(left: Vector2, right: Vector2): bool {
    return left.x == right.x && left.y == right.y;
}
```

The first parameter is the left operand, the second is the right operand. Both must be specified.

## Overloadable Operators

| Operator | Description |
|----------|-------------|
| `==` | Equality |
| `!=` | Inequality |
| `<` | Less than |
| `>` | Greater than |
| `<=` | Less than or equal |
| `>=` | Greater than or equal |
| `+` | Addition |
| `-` | Subtraction |
| `*` | Multiplication |
| `/` | Division |
| `%` | Modulo |

## Structs (Full Overloading)

Structs can overload any operator in the table above:

```
struct Vector2 {
    x: float;
    y: float;

    operator +(self: Vector2, other: Vector2): Vector2 {
        return Vector2 { x: self.x + other.x, y: self.y + other.y };
    }

    operator ==(self: Vector2, other: Vector2): bool {
        return self.x == other.x && self.y == other.y;
    }

    operator <(self: Vector2, other: Vector2): bool {
        return self.x < other.x && self.y < other.y;
    }
}

var v1: Vector2 = Vector2 { x: 1, y: 2 };
var v2: Vector2 = Vector2 { x: 3, y: 4 };
var v3: Vector2 = v1 + v2;
print(v3 == Vector2 { x: 4, y: 6 });  // true
```

### Mixed-Type Operands

The left and right operand types do not need to match:

```
operator +(self: Vector2, scalar: float): Vector2 {
    return Vector2 { x: self.x + scalar, y: self.y + scalar };
}

var v: Vector2 = Vector2 { x: 1, y: 2 } + 0.5;  // (1.5, 2.5)
```

## Classes (`==` and `!=` Only)

Classes support only `==` and `!=`. The declaration omits the explicit receiver — `this` is used instead:

```
class Person {
    private name: string;
    private age: int;

    constructor(name: string, age: int) {
        this.name = name;
        this.age = age;
    }

    operator ==(other: Person): bool {
        return this.name == other.name && this.age == other.age;
    }
}

var a = new Person("Alice", 30);
var b = new Person("Alice", 30);
print(a == b);  // true
```

- `==` and `!=` on classes without an `operator ==` definition is a compile error
- No other operators (`+`, `-`, `<`, etc.) can be overloaded on classes

## Default Equality

When no `operator ==` is defined:
- Structs compare field-by-field recursively — two structs are equal if all their fields are equal
- Classes cannot use `==` or `!=` at all (compile error)

## `.equals()` Convention

The `operator ==` is separate from a regular `.equals()` method. You can define both:

```
operator ==(self: Point, other: Point): bool {
    return self.x == other.x && self.y == other.y;
}

public equals(other: Point): bool {
    return self.x == other.x && self.y == other.y;  // regular method, not special
}
```

`.equals()` is just a method following naming convention — it has no language-level connection to `==`.

## See Also

- [Structs](15-structs.md) — value-type collections with fields and methods
- [Classes](16-classes.md) — reference types with inheritance, virtual methods, and interfaces
