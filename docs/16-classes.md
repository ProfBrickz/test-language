# Classes

Classes are reference types with inheritance, virtual dispatch, access control, and static members. They are heap-allocated by default, with escape analysis used when possible.

## Declaration

```
class Name {
    private id: int;
    protected name: string;
    final var species: string;

    constructor(name: string) {
        this.name = name;
        this.species = "Animal";
    }

    public toString(): string {
        return this.name;
    }

    private internalId(): int {
        return this.id;
    }

    public static count: int;

    public static getCount(): int {
        return Animal.count;
    }

    public virtual makeSound(): string {
        return "...";
    }

    public final identify(): string {
        return "Animal";
    }
}
```

- Members without an access modifier default to `private`
- Methods use `method` keyword (or omit when a modifier is present)
- Fields use `var` keyword (or omit when a modifier is present)
- `this` is implicit inside all methods and constructors — refers to the current instance
- Inside a class body, `method` replaces the top-level `function` keyword

### Field Declarations

Field without modifier (requires `var`):

```
var name: string;
```

Field with modifier (`var` is optional):

```
private name: string;
public var age: int;
```

### Method Declarations

Method without modifier (requires `method`):

```
method speak(): string {
    return "...";
}
```

Method with modifier (`method` is optional):

```
public speak(): string {
    return "...";
}

private doSomething(): void {
    // ...
}
```

- Methods have no explicit receiver parameter — `this` is automatically bound
- The return type is optional (defaults to `void`)

### Constructor

```
constructor(params) {
    // body
}
```

- Named `constructor` — not `method` or `function`
- Can have access modifiers: `private constructor()`, `public constructor()`
- Supports overloading: multiple constructors with different parameter lists
- `super(args)` calls the parent constructor
- `this(args)` chains to another constructor in the same class

## Access Modifiers

| Modifier | Accessible in          | Default for |
|----------|------------------------|-------------|
| private  | Same class only        | Fields and methods |
| protected| Class and subclasses   | |
| public   | Everywhere             | |

Rules:
- A subclass cannot reduce visibility when overriding (public -> protected -> private)
- A subclass can increase visibility (private -> protected -> public)
- If no modifier on an override, it inherits the parent method's visibility

## `this` and `super`

```
this.field        // access instance field
this.method()     // call instance method (virtual dispatch)
super.method()    // call parent's method (bypass virtual dispatch)
super()           // call parent constructor (only in constructor body)
this(args)        // call another constructor in same class
```

- `this` is available in all instance methods and constructors
- `super` provides access to the parent class
- `this` and `super` cannot be used in `static` methods

## Inheritance

```
class Animal {
    protected name: string;

    constructor(name: string) {
        this.name = name;
    }

    public virtual speak(): string {
        return "...";
    }

    public greet(): string {
        return "Hi from " + this.name;
    }
}

class Dog extends Animal {
    private breed: string;

    constructor(name: string, breed: string) {
        super(name);
        this.breed = breed;
    }

    public speak(): string {
        return "Woof!";
    }
}
```

- Single inheritance via `extends`
- Subclass inherits all non-private fields and methods
- `super(args)` must be called in constructor if parent has a non-default constructor
- Subclass constructors can chain with `this(args)`

## Virtual Dispatch

A method is virtual only if marked with the `virtual` keyword (or is `abstract`). Non-virtual methods are resolved at compile time against the static type.

```
class Animal {
    public virtual speak(): string { return "..."; }
    public greet(): string { return "Hi"; }
}

class Dog extends Animal {
    public speak(): string { return "Woof!"; }
    // greet() cannot be overridden — it is not virtual
}

var a: Animal = new Dog("Rex");
a.speak();     // "Woof!" — virtual dispatch to Dog's version
a.greet();     // "Hi" — no dispatch, runs Animal's version
```

- `virtual` marks a method as overridable in subclasses
- Override by redeclaring with the same name and signature
- `final` on a method prevents further overriding
- `super.method()` bypasses virtual dispatch

## Abstract Classes

```
abstract class Shape {
    public abstract area(): float;
    private abstract draw(): void;

    public describe(): string {
        return "Area: " + this.area();
    }
}

class Circle extends Shape {
    private radius: float;

    constructor(radius: float) {
        super();
        this.radius = radius;
    }

    public area(): float {
        return 3.14159 * this.radius * this.radius;
    }

    private draw(): void {
        // implementation
    }
}
```

- `abstract class` cannot be instantiated with `new`
- `abstract function` has no body — ends with `;`
- Abstract methods are implicitly `virtual`
- Concrete subclasses must implement all inherited abstract methods
- Abstract methods can have any access modifier (private, protected, public)

## Interfaces

```
interface Printable {
    public print(): void;
}

interface Serializable {
    public serialize(): string;
}

class Document extends Base implements Printable, Serializable {
    public print(): void {
        // ...
    }

    public serialize(): string {
        // ...
    }
}
```

- `interface` declares method signatures only — no fields, no constructor
- All interface methods are implicitly `public` and `abstract`
- No access modifiers on interface methods
- A class can implement multiple interfaces
- Interface methods are implicitly `virtual`
- An interface type variable can hold any implementing class instance

## Static Members

```
class Counter {
    private static count: int;

    public static increment(): void {
        Counter.count = Counter.count + 1;
    }

    public static getCount(): int {
        return Counter.count;
    }
}

Counter.increment();
Counter.increment();
print(Counter.getCount());  // 2
```

- `static` fields exist once on the class, not per instance
- `static` methods cannot use `this`
- Accessed via `ClassName.member` — not through instances
- Can combine with access modifiers: `private static count: int;`
- Static methods cannot be `virtual`, `abstract`, or `final`

## `final` and `const`

```
class Base {
    final var species: string;          // subclass cannot shadow
    const var DNA: string = "ACGT";     // truly immutable
}

final class Sealed extends Base {       // cannot be extended further
    // ...
}
```

- `final` on a class — cannot be extended
- `final` on a method — cannot be overridden (even if parent marked it virtual)
- `final` on a field — subclass cannot shadow/redeclare it
- `const` on a field — value is truly immutable after initialization (compile-time constant)
- `const` fields must have an initializer

## Instantiation: `new`

```
var dog = new Dog("Rex", "Husky");
var animal: Animal = new Dog("Fido", "Poodle");
```

- `new` keyword is required
- Calls the constructor with the given arguments
- Returns a reference to a heap-allocated object
- `new AbstractClass()` is a compile error

## Type Identity and Comparison

```
var a = new Dog("Rex", "Husky");
var b = new Dog("Rex", "Husky");
var c = a;

print(a is c);            // true — same object (reference identity)
print(a is b);            // false — different objects
print(a instanceOf Dog);  // true — type check with inheritance
print(a instanceOf Animal); // true — Dog extends Animal
// print(a == b);        // error — == not supported on class types
```

- `is` — reference identity (same heap object)
- `instanceOf` — type check including inheritance (true if subclass)
- `==` is not defined for class types — use a custom `.equals()` method for value comparison

### Type Narrowing

```
function handle(a: Animal): void {
    // a.bark();        // error — bark not on Animal

    if (a instanceOf Dog) {
        a.bark();       // ok — narrowed to Dog in this branch
    }
}
```

After an `instanceOf` check, the static type is narrowed within the branch body.

## `.equals()` Method

```
class Person {
    private name: string;
    private age: int;

    constructor(name: string, age: int) {
        this.name = name;
        this.age = age;
    }

    public equals(other: Person): bool {
        return this.name == other.name && this.age == other.age;
    }
}
```

- Classes must implement their own `.equals()` for value comparison
- No default `==` exists for class instances

## Value Semantics

Classes have reference semantics — assignment copies the reference, not the object data:

```
var a = new Dog("Rex", "Husky");
var b = a;
b.name = "Max";

print(a.name);  // "Max" — a and b share the same object
print(a is b);  // true
```

Use structs for value-type behavior with automatic copying.

## Restrictions

- No multiple inheritance (but multiple interface implementation is allowed)
- No operator overloading
- No destructors (Go garbage collector handles cleanup)
- No anonymous classes
- Classes cannot be forward-declared (must be declared before use)
- `==` cannot be used on class instances
- Methods cannot be `static` and `virtual` simultaneously
- `instanceOf` works on all types, including primitives, structs, and classes

## See Also

- [Structs](15-structs.md) — value-type alternative to classes
- [Functions](11-functions.md) — top-level function declaration
