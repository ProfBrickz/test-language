# Properties, Getters, and Setters

Properties are members with custom getter and setter accessors. Both structs and classes support properties; the main difference is that classes support access modifiers and `init` accessors, while structs do not.

For both types, every property has an implicit backing field unless it is purely calculated.

## Inline Properties

Declare a property with `{ get; set; }` after the type. Bodies are optional — omitting them uses the default accessor (returns the backing field, assigns to it):

```
struct Vector2 {
    x: float { get; set; }  // auto-property
    y: float { get; set; }
}

class Person {
    public name: string { get; set; }         // auto-property
    private age: int { get; set; }            // private auto-property
    public ssn: string { get; private set; }  // public get, private set
}
```

`get; set;` (no bodies) is equivalent to a plain field declaration.

### Custom Accessor Bodies

```
struct Temperature {
    celsius: float;

    fahrenheit: float {
        get { return celsius * 9 / 5 + 32; }
        set { celsius = (value - 32) * 5 / 9; }
    }
}

var t: Temperature = Temperature { celsius: 25 };
print(t.fahrenheit);  // 77
t.fahrenheit = 212;
print(t.celsius);     // 100
```

```
class BankAccount {
    private balance: float { get; set; }

    public formatted: string {
        get { return "$" + this.balance; }
    }

    public minimum: float {
        get { return this.minimum; }
        set {
            if (value < 0) error("Minimum cannot be negative");
            this.minimum = value;
        }
    }
}
```

- `value` is the implicit setter parameter, typed to match the property
- A property with only `get` and no `set` is read-only

## Access Modifiers (Class-Only)

In classes, accessor-level narrowing allows an accessor to be more restricted than the property itself:

```
public name: string { get; private set; }  // anyone reads, class-only writes
protected age: int { get; }                // read-only to subclasses
```

- The property-level modifier is the outer bound; accessors can narrow but not widen it
- `public x { protected get; private set; }` — accessor levels must not exceed the property level
- If an accessor has the same level as the property, the modifier can be omitted
- Structs have no access modifiers — all accessors are public

## `init` Accessor (Class-Only)

An `init` accessor is like `set`, but only callable during construction:

```
class Person {
    public name: string { get; init; }
    public age: int { get; init; }
}

var p = new Person { name = "Alice", age = 30 };
p.name = "Bob";  // error — init only, construction is done
```

Useful for immutable properties after construction while allowing convenient initializer syntax.

## Separate Accessor Declarations

Getters and setters can be declared separately from the property:

```
struct Point {
    x: int;

    get x { return x; }       // overrides default getter
    set x { x = value > 0 ? value : 0; }  // overrides default setter
}
```

```
class Foo {
    private x: int { public get; private set; }

    get x { return x; }    // overrides default getter
    set x { x = value; }   // overrides default setter
}
```

Rules:
- `get name { ... }` without a prior property declaration creates a calculated getter
- `set name { ... }` without a prior property creates a write-only property (unusual)
- Duplicate declarations for the same accessor are an error
- If a separate accessor has no custom body and the same access level as the inline declaration, a warning is given (it is redundant)

## Calculated Members

A getter alone (no backing field) is a calculated read-only member:

```
struct Circle {
    radius: float;

    get diameter: float { return radius * 2; }
    get area { return 3.14159 * radius * radius; }  // type inferred
}

class Rectangle {
    public width: float { get; init; }
    public height: float { get; init; }

    public get area: float { return this.width * this.height; }
}
```

The type can be inferred from the return expression:

```
get doubled { return value * 2; }
```

Calculated members have no setter and cannot appear in literal or initializer expressions.

A standalone setter (write-only) is allowed but unusual:

```
set callback { registerHandler(value); }
```

## See Also

- [Structs](15-structs.md) — value-type collections with fields and methods
- [Classes](16-classes.md) — reference types with inheritance, virtual methods, and interfaces
