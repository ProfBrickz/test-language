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

Condition must be a `bool`. The `else` and `else if` clauses are optional. Blocks (`{ }`) are required.

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
