# Union Types

Union types allow a variable to hold values of different types, declared with the `|` operator:

```
var x: int | string = 42;
print(x);  // 42

x = "hello";
print(x);  // "hello"

var y: int | float | bool = true;
```

## Restrictions

Type parameters are not supported on individual union members — each member is a bare type:

```
// Valid:
var x: int | float | string = 42;

// Invalid:
// var x: int{size: 32} | float{size: 64} = 42;
```

## Assignment

The value's runtime type becomes one of the union's member types:

```
var val: int | string = "hello";
print(val is string);  // true
```

## Checking with `is`

Use the `is` operator to determine which type a union variable holds:

```
var val: int | string | float = 3.14;
if (val is int) {
    print("integer");
} else if (val is string) {
    print("string");
} else if (val is float) {
    print("float");  // runs
}
```

## Binary Operations

The operator must be valid for all possible type combinations:

```
var a: int | float = 10;
var b: int | string = "hello";
// a + b: error — cannot add int/float and string
```

## Unions with Structs

Struct fields can have union types:

```
struct Config {
    name: string;
    value: int | string | float;
}

var c: Config = Config { name: "timeout", value: 30 };
c.value = "30s";
```

Union-typed struct fields follow the same narrowing rules — use `is` before accessing type-specific fields or methods. Structs themselves can also be union members:

```
var shape: Rectangle | Circle = Rectangle { ... };
if (shape is Circle) {
    print(shape.area());
}
```

## Unions with Classes

Union types work with classes and interfaces too. A union can include class types from the same hierarchy or unrelated classes:

```
var pet: Dog | Cat = new Dog("Rex");
if (pet is Dog) {
    pet.bark();
}
```

Methods on union-typed class values can only call members common to all union variants (checked at compile time):

```
class Dog { public speak(): string { return "Woof"; } }
class Cat { public speak(): string { return "Meow"; } }

var pet: Dog | Cat = new Dog("Rex");
print(pet.speak());   // OK — both Dog and Cat define speak()
// pet.bark();        // error — Cat has no bark()
```

Use `is` narrowing to access type-specific members:

```
if (pet is Dog) {
    pet.bark();       // OK — narrowed to Dog
}
```

## Next

Learn about [Structs](15-structs.md).


