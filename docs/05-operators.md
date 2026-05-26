# Operators

## Arithmetic

```
var x: int = 10;
var y: int = 3;

print(x + y);  // 13
print(x - y);  // 7
print(x * y);  // 30
print(x / y);  // 3 (integer division)
print(x % y);  // 1 (modulo)

var a: float = 10.0;
var b: float = 3.0;
print(a / b);  // 3.3333333333333335
```

- Arithmetic operators work on integers and floats
- `%` (modulo) requires integer operands
- Division and modulo by zero are runtime errors

## Comparison

```
print(5 == 5);   // true
print(5 != 3);   // true
print(3 < 5);    // true
print(5 > 3);    // true
print(3 <= 3);   // true
print(5 >= 3);   // true
```

- `==` and `!=` work on integers, floats, booleans, and null
- `<`, `>`, `<=`, `>=` work on integers and floats only (not booleans)

### Null Comparisons

```
print(null == null);   // true
print(null != null);   // false
print(null == 5);      // false
print(null != 5);      // true
```

## Logical

```
print(true && true);     // true
print(true || false);    // true
print(!true);            // false
```

- `&&` (AND), `||` (OR), `!` (NOT) work on booleans only
- Short-circuit: `false && x` returns `false` without evaluating `x`
- Short-circuit: `true || x` returns `true` without evaluating `x`
- `null` is falsy: `!null` → `true`

## Precedence

| Level | Operators |
|-------|-----------|
| 1 (highest) | `!` (unary) |
| 2 | `*`, `/`, `%` |
| 3 | `+`, `-` |
| 4 | `<`, `>`, `<=`, `>=` |
| 5 | `==`, `!=` |
| 6 | `&&` |
| 7 (lowest) | `||` |

```
print(1 + 2 * 3);       // 7
print((1 + 2) * 3);     // 9
print(true || false && false);  // true (&& binds tighter)
```

## Next

Learn about [Strings](06-strings.md).
