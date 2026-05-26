# Loops

## For Loops

C-style for loop. All three parts are optional.

```
for (init; condition; update) {
    body;
}
```

```
for (var i: int{size: 32} = 0; i < 5; i++) {
    print(i);  // 0, 1, 2, 3, 4
}

for (;;) {       // infinite loop
    break;
}

for (; i < 10;) {   // condition only (while-like)
    // ...
}
```

- Init can be a `var` declaration, `ref` declaration, or assignment
- Condition must be a `bool`
- Update can be assignment, `++`, `--`, or `%=`

### Scoping

Variables declared in the init are scoped to the entire loop. Variables declared in the body are scoped per iteration.

```
for (var i: int{size: 32} = 0; i < 3; i++) {
    var x: int{size: 32} = i * 10;
    print(x);
}
print(i);  // error: i is not defined
print(x);  // error: x is not defined
```

## For-In Loops

Iterates over elements:

```
for (var elem in iterable) {
    body;
}
```

Works on arrays, lists, and strings. For strings, each element is a single-character string.

```
var arr: array{size: 3}<int> = [10, 20, 30];
for (var x in arr) {
    print(x.toString());
}
```

## For-At Loops

Iterates over indices (0-based integers):

```
for (var index at iterable) {
    body;
}
```

Works on arrays, lists, and strings.

```
var arr: array{size: 3}<int> = [10, 20, 30];
for (var i at arr) {
    print(arr[i].toString());
}
```

## For-Of Loops

Iterates over both index and value:

```
for (var index, value of iterable) {
    body;
}
```

The first variable receives the index (integer), the second receives the value.

## While Loops

```
while (condition) {
    body;
}
```

Condition must be a `bool`. Each iteration creates a new scope.

## Break / Skip

```
break;   // exits the current loop
skip;    // skips to the next iteration (like continue)
```

Both must appear inside a loop. Break exits only the innermost loop.

```
for (var i: int{size: 32} = 0; i < 10; i++) {
    if (i == 3) {
        break;  // stops at 3
    }
    print(i);  // 0, 1, 2
}

for (var i: int{size: 32} = 0; i < 5; i++) {
    if (i == 2) {
        skip;  // skip printing 2
    }
    print(i);  // 0, 1, 3, 4
}
```

## Next

Learn about [Functions](11-functions.md).
