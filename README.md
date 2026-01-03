[![GoDoc](https://godoc.org/github.com/KarpelesLab/replvar?status.svg)](https://godoc.org/github.com/KarpelesLab/replvar)

# replvar

A Go library for variable replacement and expression evaluation in strings. It parses template strings containing `{{variable}}` expressions and resolves them against a context, supporting arithmetic operations, logical operators, comparisons, and nested field access.

## Installation

```bash
go get github.com/KarpelesLab/replvar
```

## Features

- Variable substitution with `{{name}}` syntax
- Field/member access with dot notation: `{{obj.field}}`
- Arithmetic operators: `+`, `-`, `*`, `/`, `%` (modulo)
- Bitwise operators: `|`, `&`, `^`, `~` (NOT), `<<`, `>>` (shifts)
- Logical operators: `||`, `&&`, `!`
- Comparison operators: `==`, `!=`, `<`, `<=`, `>`, `>=`
- Proper operator precedence (e.g., `2 + 3 * 4` = `14`)
- String literals with single quotes, double quotes, or backticks
- Escape sequences in double-quoted strings (`\n`, `\t`, `\r`, `\v`, `\\`)
- JSON mode for automatic JSON encoding of embedded values
- Static value optimization (compile once, resolve many times)

## Usage

### Basic Variable Replacement

```go
package main

import (
    "context"
    "fmt"

    "github.com/KarpelesLab/replvar"
)

func main() {
    // Create a context with values
    ctx := context.Background()
    ctx = context.WithValue(ctx, "name", "World")

    // Replace variables in a string
    result, err := replvar.Replace(ctx, "Hello {{name}}!", "text")
    if err != nil {
        panic(err)
    }
    fmt.Println(result) // Output: Hello World!
}
```

### Field Access

```go
ctx := context.Background()
ctx = context.WithValue(ctx, "user", map[string]any{
    "name": "Alice",
    "age":  30,
})

result, _ := replvar.Replace(ctx, "Name: {{user.name}}, Age: {{user.age}}", "text")
fmt.Println(result) // Output: Name: Alice, Age: 30
```

### Arithmetic Operations

```go
ctx := context.Background()
ctx = context.WithValue(ctx, "price", 100)
ctx = context.WithValue(ctx, "quantity", 5)

result, _ := replvar.Replace(ctx, "Total: {{price * quantity}}", "text")
fmt.Println(result) // Output: Total: 500
```

### Comparisons and Logic

```go
ctx := context.Background()
ctx = context.WithValue(ctx, "score", 85)

result, _ := replvar.Replace(ctx, "Pass: {{score == 85}}", "text")
fmt.Println(result) // Output: Pass: 1

result, _ = replvar.Replace(ctx, "High: {{score != 0}}", "text")
fmt.Println(result) // Output: High: 0
```

### Parse Once, Resolve Many

For better performance when resolving the same template multiple times:

```go
// Parse the template once
template, err := replvar.ParseString("Hello {{name}}!", "text")
if err != nil {
    panic(err)
}

// Check if the template contains only static values
if template.IsStatic() {
    // No variables, result is constant
}

// Resolve with different contexts
ctx1 := context.WithValue(context.Background(), "name", "Alice")
ctx2 := context.WithValue(context.Background(), "name", "Bob")

result1, _ := template.Resolve(ctx1) // "Hello Alice!"
result2, _ := template.Resolve(ctx2) // "Hello Bob!"
```

### JSON Mode

When using `"json"` mode, embedded variables are automatically JSON-encoded:

```go
ctx := context.Background()
ctx = context.WithValue(ctx, "data", map[string]any{"key": "value"})

result, _ := replvar.Replace(ctx, `{"nested": {{data}}}`, "json")
// Output: {"nested": {"key":"value"}}
```

### String Literals

Variables can contain string literals with different quote types:

```go
// Single quotes - no escape processing
result, _ := replvar.Replace(ctx, "Value: {{'hello world'}}", "text")

// Double quotes - supports escape sequences
result, _ := replvar.Replace(ctx, "Value: {{\"hello\\tworld\"}}", "text")

// Backticks - raw strings, no escape processing
result, _ := replvar.Replace(ctx, "Value: {{`hello\\nworld`}}", "text")
```

## API Reference

### Functions

#### `Replace(ctx context.Context, s string, mode string) (string, error)`

Parses and resolves a template string in one step. This is the simplest way to perform variable replacement.

- `ctx`: Context containing variable values (accessed via `ctx.Value(key)`)
- `s`: Template string with `{{variable}}` expressions
- `mode`: Either `"text"` or `"json"` (for automatic JSON encoding)

#### `ParseString(s string, mode string) (Var, error)`

Parses a template string into a `Var` that can be resolved multiple times.

#### `ParseVariable(s string) (Var, error)`

Parses a variable expression (the content inside `{{}}`).

### Var Interface

```go
type Var interface {
    Resolve(context.Context) (any, error)
    IsStatic() bool
}
```

- `Resolve`: Evaluates the expression against the given context
- `IsStatic`: Returns `true` if the value is constant (no variables)

## Expression Syntax

| Syntax | Description | Example |
|--------|-------------|---------|
| `{{name}}` | Variable lookup | `{{username}}` |
| `{{a.b}}` | Field access | `{{user.email}}` |
| `{{a + b}}` | Addition | `{{price + tax}}` |
| `{{a - b}}` | Subtraction | `{{total - discount}}` |
| `{{a * b}}` | Multiplication | `{{qty * price}}` |
| `{{a / b}}` | Division | `{{total / count}}` |
| `{{a % b}}` | Modulo | `{{index % 2}}` |
| `{{a << b}}` | Left shift | `{{1 << 4}}` |
| `{{a >> b}}` | Right shift | `{{16 >> 2}}` |
| `{{a \| b}}` | Bitwise OR | `{{flags \| mask}}` |
| `{{a & b}}` | Bitwise AND | `{{flags & mask}}` |
| `{{a ^ b}}` | Bitwise XOR | `{{a ^ b}}` |
| `{{~a}}` | Bitwise NOT | `{{~mask}}` |
| `{{a \|\| b}}` | Logical OR | `{{a \|\| b}}` |
| `{{a && b}}` | Logical AND | `{{a && b}}` |
| `{{!a}}` | Logical NOT | `{{!enabled}}` |
| `{{a == b}}` | Equality | `{{status == 'ok'}}` |
| `{{a != b}}` | Inequality | `{{status != 'error'}}` |
| `{{a < b}}` | Less than | `{{age < 18}}` |
| `{{a <= b}}` | Less than or equal | `{{score <= 100}}` |
| `{{a > b}}` | Greater than | `{{count > 0}}` |
| `{{a >= b}}` | Greater than or equal | `{{level >= 5}}` |
| `{{'str'}}` | Single-quoted string | `{{'hello'}}` |
| `{{"str"}}` | Double-quoted string (with escapes) | `{{"hello\n"}}` |
| `` {{`str`}} `` | Backtick string (raw) | `` {{`hello`}} `` |
| `{{123}}` | Number literal | `{{42}}` |
| `{{1.5}}` | Float literal | `{{3.14}}` |

## Operator Precedence

Operators are evaluated according to standard precedence rules (higher precedence binds tighter):

| Precedence | Operators | Description |
|------------|-----------|-------------|
| 1 (highest) | `.` | Member access |
| 2 | `!` `~` | Unary NOT (logical, bitwise) |
| 3 | `*` `/` `%` | Multiplication, division, modulo |
| 4 | `+` `-` | Addition, subtraction |
| 5 | `<<` `>>` | Bit shifts |
| 6 | `<` `<=` `>` `>=` | Relational comparisons |
| 7 | `==` `!=` | Equality comparisons |
| 8 | `&` | Bitwise AND |
| 9 | `^` | Bitwise XOR |
| 10 | `\|` | Bitwise OR |
| 11 | `&&` | Logical AND |
| 12 (lowest) | `\|\|` | Logical OR |

For example, `2 + 3 * 4` evaluates to `14` (not `20`), and `1 || 0 && 0` evaluates to `1` (not `0`).

## License

See LICENSE file for details.
