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

Condition must be a `bool`. The `else` and `else if` clauses are optional. Braces are required for multi-statement bodies, but optional for single-statement bodies.

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

Single-statement bodies can omit braces:

```
if (condition) statement;
else if (condition) statement;
else statement;

for (init; condition; update) statement;

while (condition) statement;
```

## For Loops

```
for (init; condition; update) {
    body;
}
```

C-style for loop. All three parts are optional:

```
for (;;) {       // infinite loop (use break)
    // ...
}

for (; i < 10;) {   // condition only (while-like)
    // ...
}
```

The init can be a `var` declaration or an assignment. The condition must be a `bool`. The update can be an assignment, `++`, or `--`.

The init, condition, and update share a scope. Each iteration creates a new scope for the body. Variables declared in the body are not visible in the condition or update.

## For-In Loops

```
for (var elem in iterable) {
    body;
}
```

Iterates over the elements of an array, list, or string. For strings, each element is a single-character string.

```
var arr: array{size: 3}<int> = [10, 20, 30];
for (var x in arr) {
    print(x.toString());
}
// Output:
// 10
// 20
// 30
```

Each iteration creates a new scope containing the loop variable.

## For-At Loops

```
for (var index at iterable) {
    body;
}
```

Iterates over the indices (0-based integers) of an array, list, or string.

```
var arr: array{size: 3}<int> = [10, 20, 30];
for (var i at arr) {
    print(arr[i].toString());
}
// Output:
// 10
// 20
// 30
```

Each iteration creates a new scope containing the loop variable.

## For-Of Loops

```
for (var index, value of iterable) {
    body;
}
```

Iterates over both the index and value of an array, list, or string. The first variable receives the index (integer), the second receives the value (element or single-character string).

```
var arr: array{size: 3}<int> = [10, 20, 30];
for (var i, v of arr) {
    print((i).toString());
    print((v).toString());
}
// Output:
// 0
// 10
// 1
// 20
// 2
// 30
```

Each iteration creates a new scope containing both loop variables.

## While Loops

```
while (condition) {
    body;
}
```

Condition must be a `bool`. Each iteration creates a new scope for the body.

## Break / Skip

```
break;
skip;
```

`break` exits the current loop. `skip` skips to the next iteration (like `continue` in C). Both must appear inside a loop. Using them outside a loop is a runtime error.

Break exits the innermost loop only:

```
for (var i: int{size: 32} = 0; i < 3; i++) {
    for (var j: int{size: 32} = 0; j < 3; j++) {
        if (j == 1) {
            break;  // breaks inner loop only
        }
    }
}
```

## Scoping Rules

All control flow bodies (`if`, `else`, `for`, `while`) create new lexical scopes:

- Variables declared inside a block are not visible outside
- Outer variables are accessible from within blocks
- Assignments to outer variables modify the outer variable (not a shadow)

For for loops specifically:

- Variables declared in the `init` are scoped to the entire loop (init, condition, update, body)
- Variables declared in the `body` are scoped per iteration (fresh scope each iteration)
- Init variables are NOT accessible after the loop ends

```
for (var i: int{size: 32} = 0; i < 3; i++) {
    var x: int{size: 32} = i;  // scoped to this iteration
    print(x);
}
print(i);  // error: i is not defined
print(x);  // error: x is not defined
```
