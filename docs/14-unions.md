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
print(val instanceOf string);  // true
```

## Checking with `instanceOf`

Use the `instanceOf` operator to determine which type a union variable holds:

```
var val: int | string | float = 3.14;
if (val instanceOf int) {
    print("integer");
} else if (val instanceOf string) {
    print("string");
} else if (val instanceOf float) {
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

## Next

Learn about [Structs](15-structs.md).


