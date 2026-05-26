# Strings

The `string` type holds UTF-8 text.

## Declaration

```
var s: string = "hello";
var t: string{size: 5} = "world";       // fixed-size: exactly 5 chars
var u: string{min: 1, max: 10} = "hi";  // constrained: 1 to 10 chars
```

Parameters:
- **size** — exact character count (mutually exclusive with min/max)
- **min** — minimum characters
- **max** — maximum characters

## String Literals

```
var a: string = "hello";
var b: string = "hello\nworld";  // with escape
```

Escape sequences:

| Sequence | Result |
|----------|--------|
| `\n` | Newline |
| `\t` | Tab |
| `\\` | Backslash |
| `\"` | Double quote |
| `\'` | Single quote |
| `\r` | Carriage return |
| `\0` | Null byte |
| `\xNN` | Hex byte |
| `\uNNNN` | Unicode (4 hex digits) |
| `\UNNNNNNNN` | Unicode (8 hex digits) |

## Concatenation

```
var s: string = "hello " + "world";  // "hello world"
var t: string = "count: " + 42;      // "count: 42"
```

The `+` operator concatenates strings. Non-string operands are implicitly converted to strings.

## Indexing

```
var s: string = "hello";
print(s[0]);  // "h"
print(s[4]);  // "o"
```

Index bounds are checked at runtime.

## Methods

### `.length`

Returns the number of characters (runes):

```
print(s.length);  // 5
```

### `.concat(s)`

Returns a new string with `s` appended:

```
var t: string = s.concat(" world");
```

### `.add(ch)` / `.add(ch, index)`

Inserts a single character (used as a statement, mutates the variable):

```
var s: string = "hello";
s.add("!");
print(s);  // "hello!"
```

Cannot be used on fixed-size strings.

### `.remove(index)`

Removes and returns the character at `index`:

```
var c: string = s.remove(0);
print(c);  // "h"
print(s);  // "ello"
```

Cannot be used on fixed-size strings.

### `.toString()`

Every string value has `.toString()` returning itself.

## Next

Learn about [Type Conversion](07-type-conversion.md).
