go-peg
======

Go language [PEG](http://en.wikipedia.org/wiki/Parsing_expression_grammar) (Parsing Expression Grammars) library.

This library is ported from [cpp-pegib](https://github.com/yhirose/cpp-peglib).

```go
package main

import (
	"fmt"
	. "github.com/yhirose/go-peg"
	"strconv"
)

func main() {
	// Create a PEG parser
	parser, _ := NewParser(`
        # Grammar for simple calculator...
        EXPRESSION       <-  TERM (TERM_OPERATOR TERM)*
        TERM             <-  FACTOR (FACTOR_OPERATOR FACTOR)*
        FACTOR           <-  NUMBER / '(' EXPRESSION ')'
        TERM_OPERATOR    <-  [-+]
        FACTOR_OPERATOR  <-  [/*]
        NUMBER           <-  [0-9]+
    `)

	// Setup actions
	reduce := func(sv *SemanticValues, dt Any) (Any, error) {
		ret := sv.ToInt(0)
		for i := 1; i < len(sv.Vs); i += 2 {
			num := sv.ToInt(i + 1)
			switch sv.ToStr(i) {
			case "+":
				ret += num
			case "-":
				ret -= num
			case "*":
				ret *= num
			case "/":
				ret /= num
			}
		}
		return ret, nil
	}

	g := parser.Grammar

	g["EXPRESSION"].Action = reduce
	g["TERM"].Action = reduce
	g["TERM_OPERATOR"].Action = func(sv *SemanticValues, dt Any) (Any, error) { return sv.S, nil }
	g["FACTOR_OPERATOR"].Action = func(sv *SemanticValues, dt Any) (Any, error) { return sv.S, nil }
	g["NUMBER"].Action = func(sv *SemanticValues, dt Any) (Any, error) { return strconv.Atoi(sv.S) }

	// Parse
	if val, err := parser.ParseAndGetValue("1+2*3*(4-5+6)/7-8"); err == nil {
		fmt.Println(val) // -3
	} else {
		fmt.Println(err)
	}
}
```

TODO
----
 * Memoization (Packrat parsing)
 * Skip whitespaces
 * AST generation

License
-------

MIT license (© 2016 Yuji Hirose)