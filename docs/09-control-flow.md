# Control Flow

## If / Else

```
if (condition) {
    statements;
} else if (condition) {
    statements;
} else {
    statements;
}
```

- Condition must be a `bool`
- `else` and `else if` are optional
- Braces required for multi-statement bodies, optional for single-statement bodies

### Single-Statement Bodies

```
if (condition) statement;
else if (condition) statement;
else statement;
```

## Scoping

Each `if`, `else if`, and `else` body creates a new scope:

```
var x: int{size: 32} = 1;

if (true) {
    var y: int{size: 32} = 2;  // scoped to this block
    x = 3;                      // modifies outer x
}

print(x);  // 3
print(y);  // error: y is not defined
```

Outer variables are accessible and modifiable from within blocks:

```
var x: int{size: 32} = 0;

if (true) {
    x = 5;  // modifies outer x
}

print(x);  // 5
```

## Switch

The `switch` statement evaluates an expression and matches against `case` clauses. It does **not** fall through.

```
switch (expression) {
    case (== value) {
        body;
    }
    case (< value) {
        body;
    }
    default {
        body;
    }
}
```

### Case Operators

Each case can specify a relational operator. If omitted, `==` is used.

| Operator | Meaning    |
|----------|------------|
| `==`     | Equal to   |
| `!=`     | Not equal  |
| `<`      | Less than  |
| `>`      | Greater than |
| `<=`     | Less than or equal |
| `>=`     | Greater than or equal |

```
var x: int{size: 32} = 5;

switch (x) {
    case (== 0) {
        print("zero");
    }
    case (< 5) {
        print("less than 5");
    }
    case (5) {
        print("five");  // runs
    }
}
```

### Default

The `default` case matches if no other case matched:

```
switch (x) {
    case (== 0) {
        print("zero");
    }
    default {
        print("something else");
    }
}
```

### No Fallthrough

Only the first matching case executes:

```
switch (x) {
    case (== 5) {
        print("five");
    }
    case (== 5) {
        print("five again");  // never reached
    }
}
```

### Scoping

Each case body creates a new scope:

```
switch (x) {
    case (== 1) {
        var msg: string = "one";
        print(msg);
    }
}
```

## Next

Learn about [Loops](10-loops.md).
