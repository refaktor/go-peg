go-peg
======

**NOTE**: This library is **deprecated**. Please use [**cpp-peglib**](https://github.com/yhirose/cpp-peglib) instead.

Yet another [PEG](http://en.wikipedia.org/wiki/Parsing_expression_grammar) (Parsing Expression Grammars) parser generator for Go.

If you need a PEG grammar checker, you may want to check [**peglint**](https://github.com/yhirose/go-peg/tree/master/cmd/peglint).

If you need a C++ version, please see [*cpp-peglib*](https://github.com/yhirose/cpp-peglib).

### Extended features

 * Token operator: `<` `>`
 * Automatic whitespace skipping: `%whitespace`
 * Expression parsing for binary operators ([precedence climbing method](https://en.wikipedia.org/wiki/Operator-precedence_parser#Precedence_climbing_method))
 * Parameterized rule or Macro
 * Word expression: `%word`
 * AST generation
 * Enhanced error reporting with suggestions
 * Error recovery during parsing
 * Detailed tracing for debugging

### Usage

```go
// Create a PEG parser
parser, _ := NewParser(`
    # Simple calculator
    EXPR         ←  ATOM (BINOP ATOM)*
    ATOM         ←  NUMBER / '(' EXPR ')'
    BINOP        ←  < [-+/*] >
    NUMBER       ←  < [0-9]+ >
    %whitespace  ←  [ \t]*
    ---
    # Expression parsing option
    %expr  = EXPR   # Rule to apply 'precedence climbing method' to
    %binop = L + -  # Precedence level 1
    %binop = L * /  # Precedence level 2
`)

// Setup semantic actions
g := parser.Grammar
g["EXPR"].Action = func(v *Values, d Any) (Any, error) {
    val := v.ToInt(0)
    if v.Len() > 1 {
        ope := v.ToStr(1)
        rhs := v.ToInt(2)
        switch ope {
        case "+": val += rhs
        case "-": val -= rhs
        case "*": val *= rhs
        case "/": val /= rhs
        }
    }
    return val, nil
}
g["BINOP"].Action = func(v *Values, d Any) (Any, error) {
    return v.Token(), nil
}
g["NUMBER"].Action = func(v *Values, d Any) (Any, error) {
    return strconv.Atoi(v.Token())
}

// Parse
input := " 1 + 2 * 3 * (4 - 5 + 6) / 7 - 8 "
val, _ := parser.ParseAndGetValue(input, nil)

fmt.Println(val) // Output: -3
```

Parameterized Rule or Macro
---------------------------

```peg
# Syntax
Start      ← _ Expr
Expr       ← Sum
Sum        ← List(Product, SumOpe)
Product    ← List(Value, ProOpe)
Value      ← Number / T('(') Expr T(')')

# Token
SumOpe     ← T('+' / '-')
ProOpe     ← T('*' / '/')
Number     ← T([0-9]+)
~_         ← [ \t\r\n]*

# Macro
List(I, D) ← I (D I)*
T(x)       ← < x > _
```

Word expression
---------------

```go
parser, _ := NewParser(`
    ROOT         ←  'hello' 'world'
    %whitespace  ←  [ \t\r\n]*
    %word        ←  [a-z]+
`)

parser.Parse("hello world", nil) # OK
parser.Parse("helloworld", nil)  # NG
```

AST generation
--------------

```go
// Create a PEG parser
parser, _ := NewParser(`
    EXPRESSION       <-  TERM (TERM_OPERATOR TERM)*
    TERM             <-  FACTOR (FACTOR_OPERATOR FACTOR)*
    FACTOR           <-  NUMBER / '(' EXPRESSION ')'
    TERM_OPERATOR    <-  < [-+] >
    FACTOR_OPERATOR  <-  < [/*] >
    NUMBER           <-  < [0-9]+ >
    %whitespace      <-  [ \t\r\n]*
`)

// Evaluator
var eval func(ast *Ast) int
eval = func(ast *Ast) int {
    if ast.Name == "NUMBER" {
        val, _ := strconv.Atoi(ast.Token)
        return val
    } else {
        nodes := ast.Nodes
        val := eval(nodes[0])
        for i := 1; i < len(nodes); i += 2 {
            num := eval(nodes[i+1])
            ope := nodes[i].Token[0]
            switch ope {
            case '+':
                val += num
                break
            case '-':
                val -= num
                break
            case '*':
                val *= num
                break
            case '/':
                val /= num
                break
            }
        }
        return val
    }
}

// Generate AST
parser.EnableAst()
input := " 1 + 2 * 3 * (4 - 5 + 6) / 7 - 8 "
ret, _ := parser.ParseAndGetValue(input, nil)
ast := ret.(*Ast)

// Optimize AST
opt := NewAstOptimizer(nil)
ast = opt.Optimize(ast, nil)

// Evaluate AST
val := eval(ast)

fmt.Println(val) // Output: -3
```

Error Reporting and Recovery
---------------------------

The library provides enhanced error reporting with detailed context and suggestions:

```go
parser, _ := NewParser(grammar)
_, err := parser.ParseAndGetValue(input, nil)
if err != nil {
    fmt.Println(err) // Shows detailed error with line/column and pointer
    
    // For syntax errors, you can get suggestions
    if syntaxErr, ok := err.(*SyntaxError); ok {
        for _, suggestion := range syntaxErr.GetSuggestions() {
            fmt.Println("- " + suggestion)
        }
    }
}
```

You can also enable error recovery to continue parsing after errors:

```go
parser, _ := NewParser(grammar)
parser.RecoveryEnabled = true
parser.MaxErrors = 10 // Maximum number of errors to report
val, errs := parser.ParseAndGetValueWithRecovery(input, nil)
for _, err := range errs {
    fmt.Println(err)
}
```

Tracing for Debugging
--------------------

Enable detailed tracing to debug your grammar:

```go
parser, _ := NewParser(grammar)
parser.EnableTracing(&TracingOptions{
    ShowRuleEntry:    true,
    ShowRuleExit:     true,
    ShowTokens:       false,
    ShowErrorContext: true,
    OutputFormat:     "text",
})
```

Using peglint
------------

The `peglint` tool now supports error recovery and better error reporting:

```bash
# Basic usage
peglint grammar.peg -f source.txt

# With error recovery
peglint -recovery -max-errors 5 grammar.peg -f source.txt

# With tracing
peglint -trace grammar.peg -f source.txt
```

TODO
----

 * Memoization (Packrat parsing)

License
-------

MIT license (© 2016 Yuji Hirose)
