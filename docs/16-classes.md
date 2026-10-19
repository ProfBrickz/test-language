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
- When an access modifier is present, `method` and `var` are implied — `public speak()` is parsed as `public method speak()`
- When no modifier is present, `method` and `var` are required
- `this` is implicit inside all methods and constructors — refers to the current instance
- Inside a class body, `method` replaces the top-level `function` keyword

### Field Declarations

Field without modifier — `var` keyword required:

```
var name: string;
```

Field with modifier — `var` is implied, not written:

```
private name: string;
public age: int;
```

Same rule as methods — the modifier makes `var` redundant:

```
public var age: int;   // valid but redundant — use public age: int instead
```

### Method Declarations

Method without modifier — `method` keyword required:

```
method speak(): string {
    return "...";
}
```

Method with modifier — `method` is implied, not written:

```
public speak(): string {
    return "...";
}

private doSomething(): void {
    // ...
}
```

The access modifier makes the declaration's intent clear — the `method` keyword would be redundant:

```
public method speak(): string {   // valid but redundant — use public speak() instead
    return "...";
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

| Modifier  | Accessible in        | Default for        |
| --------- | -------------------- | ------------------ |
| private   | Same class only      | Fields and methods |
| protected | Class and subclasses |                    |
| public    | Everywhere           |                    |

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
    protected abstract draw(): void;

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

    protected draw(): void {
        // implementation
    }
}
```

- `abstract class` cannot be instantiated with `new`
- `abstract function` has no body — ends with `;`
- Abstract methods are implicitly `virtual`
- Concrete subclasses must implement all inherited abstract methods
- Abstract methods can have any access modifier except `private` (subclasses must be able to see them to implement them)

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

print(a isRef c);            // true — same object (reference identity)
print(a isRef b);            // false — different objects
print(a is Dog);  // true — type check with inheritance
print(a is Animal); // true — Dog extends Animal
// print(a == b);        // error — == not supported on class types
```

- `isRef` — reference identity (same heap object)
- `is` — type check including inheritance (true if subclass)
- `==` and `!=` work only if the class defines `operator ==` (see [Operator Overloading](18-operator-overloading.md)) — otherwise they are a compile error

### Type Narrowing

```
function handle(a: Animal): void {
    // a.bark();        // error — bark not on Animal

    if (a is Dog) {
        a.bark();       // ok — narrowed to Dog in this branch
    }
}
```

After an `is` check, the static type is narrowed within the branch body.

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

- Classes can implement their own `.equals()` for value comparison, but it is a plain method — not connected to `==` or `!=`
- For `==` and `!=` support, define `operator ==` instead (see [Operator Overloading](18-operator-overloading.md))

## Value Semantics

Classes have reference semantics — assignment copies the reference, not the object data:

```
var a = new Dog("Rex", "Husky");
var b = a;
b.name = "Max";

print(a.name);  // "Max" — a and b share the same object
print(a isRef b);  // true
```

Use structs for value-type behavior with automatic copying.

## Cross-Class Access

Classes can grant access to their private or protected fields and methods to other classes using the `allow` keyword. This allows controlled access to encapsulated members without changing their visibility modifiers.

### Syntax

A class can define an `allow` block with one or more access rules:

```
class Name {
    // ...
    // Single rule forms:
    allow { OtherClass };                                          // full access to all properties
    allow { OtherClass: { fieldName } };                           // access to specific property (getter + setter)
    allow { OtherClass: { fieldName: { get } } };                  // access to getter only
    allow { OtherClass: { fieldName: { set } } };                  // access to setter only

    // Multiple rules in one block:
    allow {
        Employee: { base: { get } },
        HR: { base: { set } },
        Manager: { bonus }
    };
}
```

### Access Control Levels

Each access rule specifies what another class can access:

| Syntax                                        | Access Level                              |
| --------------------------------------------- | ----------------------------------------- |
| `{ OtherClass }`                              | Full access to all properties and methods |
| `{ OtherClass: { fieldName } }`               | Getter and setter for fieldName           |
| `{ OtherClass: { fieldName: { get } } }`      | Read-only (getter only)                   |
| `{ OtherClass: { fieldName: { set } } }`      | Write-only (setter only)                  |
| `{ OtherClass: { fieldName: { get, set } } }` | Full access to fieldName                  |

### Examples

#### Full Access to All Properties

Grant one class full access (getter and setter) to all properties:

```
class Animal {
    private id: int;
    protected name: string;

    allow { Dog };  // Dog can access all fields: id, name, etc.
}

class Dog extends Animal {
    constructor(name: string) {
        this.name = name;
        // this.id is now accessible here
        print(this.id);        // ok
        this.name = "Max";     // ok
    }
}
```

#### Access to Specific Properties

Grant one class access to specific properties only:

```
class BankAccount {
    private balance: float;
    private owner: string;

    allow { User: { owner } };  // User can only access owner field
}

class User {
    var userId: string;

    function login(account: BankAccount): void {
        // account.owner is accessible
        print(account.owner);
        // account.balance is NOT accessible
        // account.balance = 100.0;  // error
    }
}
```

#### Multiple Classes Accessing Different Properties

Multiple access rules in a single `allow` block:

```
class Salary {
    private base: float;
    private bonus: float;
    private commission: float;

    allow {
        Employee: { base: { get } },        // Employee can read base salary
        HR: { base: { set } },              // HR can modify base salary
        Manager: {
            bonus: { get, set },            // Manager can read/write bonus
            commission: { get }             // Manager can read commission
        }
    };
}

class Employee {
    function getBasePay(salary: Salary): float {
        // salary.base is readable
        return salary.base;
        // salary.base = 1000.0;  // error — setter not allowed
    }
}

class HR {
    function updateSalary(salary: Salary): void {
        // salary.base is writable
        salary.base = 1500.0;  // ok
    }
}

class Manager {
    function adjustCompensation(salary: Salary): void {
        // salary.bonus and salary.commission are readable
        print(salary.bonus);
        salary.bonus = salary.bonus + 100;  // ok
        // salary.commission = 500;  // error — setter not allowed
    }
}
```

#### Accessing Methods

The `allow` syntax also grants access to methods, not just fields:

```
class Printer {
    private data: string;

    allow { Console: { data } };  // Console can access data field
    allow { Console: { print } }; // Console can call print method
}

class Console {
    function log(printer: Printer): void {
        // printer.data is accessible
        print(printer.data);
        // printer.print("hello");  // also accessible
    }
}
```

#### Accessing Static Members

Static members are also accessible with the `allow` syntax:

```
class Config {
    private static apiKey: string;

    allow { API: { apiKey } };  // API can access the static apiKey
}

class API {
    function request(): void {
        print(Config.apiKey);  // accessible
    }
}
```

#### Complex Access Patterns

Access different levels of access for different classes:

```
class User {
    private password: string;
    private email: string;
    private profile: Profile;

    allow {
        // Admin gets full access to everything
        Admin: { password: { get, set }, email: { get, set }, profile: { get, set } },
        // Manager can read user info but not password
        Manager: { password: { get }, email: { get, set }, profile: { get } },
        // Support can only read email
        Support: { email: { get } }
    };
}

class Admin {
    function resetPassword(user: User): void {
        user.password = "newpassword";  // ok
    }
}

class Manager {
    function viewUser(user: User): void {
        print(user.email);  // ok
        // user.password = "xxx";  // error
    }
}
```

### Rules

- `allow` declarations are class-level statements inside the class body
- A class can `allow` multiple other classes in a single block
- A class can `allow` multiple properties or methods from the same class
- Access grants are not inherited — subclasses do not inherit parent's `allow` declarations
- If a property is not explicitly allowed, it remains inaccessible to other classes
- The `allow` keyword is processed at class definition time
- Access rules apply to both the class where they're defined and its instances
- Static members can be accessed via the class name (e.g., `Config.apiKey`) or through instances
- When using `allow`, you can access the allowed members using normal dot notation

### Multiple Access Rules

A single `allow` block can contain multiple access rules, granting different classes access to different properties:

```
class Base {
    private baseField: string;
    private baseMethod: () => void;

    allow {
        Employee: { baseField: { get } },
        HR: { baseField: { set } }
    };
}
```

### Full Access

Grant one class full access (getter and setter) to all properties:

```
class Animal {
    private id: int;
    protected name: string;

    allow { Dog };  // Dog can access all fields: id, name, etc.
}

class Dog extends Animal {
    constructor(name: string) {
        this.name = name;
        // this.id is now accessible here
        print(this.id);        // ok
        this.name = "Max";     // ok
    }
}
```

### Access to Specific Properties

Grant one class access to specific properties only:

```
class BankAccount {
    private balance: float;
    private owner: string;

    allow { User: { owner } };  // User can only access owner field
}

class User {
    var userId: string;

    function login(account: BankAccount): void {
        // account.owner is accessible
        print(account.owner);
        // account.balance is NOT accessible
        // account.balance = 100.0;  // error
    }
}
```

### Access to Specific Getter or Setter

Grant one class access to only the getter or only the setter:

```
class Salary {
    private base: float;
    private bonus: float;

    allow { Employee: { base: { get } } };     // Employee can read base salary
    allow { HR: { base: { set } } };          // HR can modify base salary
}

class Employee {
    function getBasePay(salary: Salary): float {
        // salary.base is readable
        return salary.base;
        // salary.base = 1000.0;  // error — setter not allowed
    }
}

class HR {
    function updateSalary(salary: Salary): void {
        // salary.base is writable
        salary.base = 1500.0;  // ok
        // salary.base = salary.base + 100;  // ok
    }
}
```

### Accessing Methods

The `allow` syntax also grants access to methods, not just fields:

```
class Printer {
    private data: string;

    allow { Console: { data } };  // Console can access data field
}

class Console {
    function log(printer: Printer): void {
        // printer.data is accessible
        print(printer.data);
    }
}
```

### Accessing Static Members

Static members are also accessible with the `allow` syntax:

```
class Config {
    private static apiKey: string;

    allow { API: { apiKey } };  // API can access the static apiKey
}

class API {
    function request(): void {
        print(Config.apiKey);  // accessible
    }
}
```

### Rules

- `allow` declarations are class-level statements inside the class body
- A class can `allow` multiple other classes
- A class can `allow` multiple properties or methods from the same class
- Access grants are not inherited — subclasses do not inherit parent's `allow` declarations
- If a property is not explicitly allowed, it remains inaccessible to other classes
- Methods can be allowed with full access or restricted to `get`/`set` semantics (for getters/setters)

## Restrictions

- No multiple inheritance (but multiple interface implementation is allowed)
- Operator overloading is limited to `==` and `!=` (via `operator ==`)
- No destructors (Go garbage collector handles cleanup)
- No anonymous classes
- Classes cannot be forward-declared (must be declared before use)
- `==` and `!=` work on classes only if `operator ==` is defined — otherwise a compile error
- Methods cannot be `static` and `virtual` simultaneously
- Abstract methods cannot be `private` (subclasses must be able to implement them)
- `is` works on all types, including primitives, structs, and classes

## See Also

- [Operator Overloading](18-operator-overloading.md) — custom operator behavior for structs and classes
- [Properties, Getters, and Setters](17-properties.md) — custom accessor logic for fields
- [Structs](15-structs.md) — value-type alternative to classes
- [Functions](11-functions.md) — top-level function declaration
