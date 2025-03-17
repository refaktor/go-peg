package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"strings"

	peg "github.com/yhirose/go-peg"
)

var usageMessage = `usage: peglint [-ast] [-opt] [-trace] [-f path] [-s string] [grammar path]

peglint checks syntax of a given PEG grammar file and reports errors. If the check is successful and a user gives a source file for the grammar, it will also check syntax of the source file.

The -ast flag prints the AST (abstract syntax tree) of the source file.

The -opt flag prints the optimized AST (abstract syntax tree) of the source file.

The -trace flag can be used with the source file. It prints names of rules and operators that the PEG parser detects on standard error.

The -f 'path' specifies a file path to the source text.

The -s 'string' specifies the source text.
`

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(1)
}

var (
	astFlag        = flag.Bool("ast", false, "show ast")
	optFlag        = flag.Bool("opt", false, "show optimized ast")
	traceFlag      = flag.Bool("trace", false, "show trace message")
	sourceFilePath = flag.String("f", "", "source file path")
	sourceString   = flag.String("s", "", "source string")
	profPath       = flag.String("prof", "", "write cpu profile to file")
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func pcheck(err error) {
	if perr, ok := err.(*peg.Error); ok {
		fmt.Println("Error details:")
		for _, d := range perr.Details {
			fmt.Println(d)
		}

		// Show suggestions if available
		suggestions := perr.GetSuggestions()
		if len(suggestions) > 0 {
			fmt.Println("\nSuggestions:")
			for _, suggestion := range suggestions {
				fmt.Println("- " + suggestion)
			}
		}

		os.Exit(1)
	} else if syntaxErr, ok := err.(*peg.SyntaxError); ok {
		fmt.Println("Syntax error details:")
		fmt.Println(syntaxErr.Error())

		// Show expected tokens and suggestions
		fmt.Println("\nExpected tokens:", strings.Join(syntaxErr.Expected, ", "))

		// Show suggestions
		suggestions := syntaxErr.GetSuggestions()
		if len(suggestions) > 0 {
			fmt.Println("\nSuggestions:")
			for _, suggestion := range suggestions {
				fmt.Println("- " + suggestion)
			}
		}

		os.Exit(1)
	}
}

func SetupTracer(p *peg.Parser) {
	// Use the new tracing options
	p.EnableTracing(&peg.TracingOptions{
		ShowRuleEntry:    true,
		ShowRuleExit:     true,
		ShowTokens:       false,
		ShowErrorContext: true,
		OutputFormat:     "text",
	})

	fmt.Println("Tracing enabled. Parser will show rule entry/exit and error context.")
}

var (
	recoveryFlag  = flag.Bool("recovery", false, "enable error recovery")
	maxErrorsFlag = flag.Int("max-errors", 10, "maximum number of errors to report")
)

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		usage()
	}

	dat, err := ioutil.ReadFile(args[0])
	check(err)

	parser, err := peg.NewParser(string(dat))
	pcheck(err)

	var source string

	if *sourceFilePath != "" {
		if *sourceFilePath == "-" {
			dat, err := ioutil.ReadAll(os.Stdin)
			check(err)
			source = string(dat)
		} else {
			dat, err := ioutil.ReadFile(*sourceFilePath)
			check(err)
			source = string(dat)
		}
	}

	if *sourceString != "" {
		source = *sourceString
	}

	if len(source) > 0 {
		if *traceFlag {
			SetupTracer(parser)
		}

		if *astFlag || *optFlag {
			parser.EnableAst()
		}

		// Enable error recovery if requested
		if *recoveryFlag {
			parser.RecoveryEnabled = true
			parser.MaxErrors = *maxErrorsFlag
			fmt.Printf("Error recovery enabled (max errors: %d)\n", parser.MaxErrors)
		}

		if *profPath != "" {
			f, err := os.Create(*profPath)
			check(err)
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}

		if *recoveryFlag {
			// Use recovery mode parsing
			val, errs := parser.ParseAndGetValueWithRecovery(source, nil)
			if len(errs) > 0 {
				fmt.Printf("Parsing completed with %d errors\n", len(errs))
				for i, err := range errs {
					fmt.Printf("\nError #%d:\n", i+1)
					if syntaxErr, ok := err.(*peg.SyntaxError); ok {
						fmt.Println(syntaxErr.Error())

						suggestions := syntaxErr.GetSuggestions()
						if len(suggestions) > 0 {
							fmt.Println("\nSuggestions:")
							for _, suggestion := range suggestions {
								fmt.Println("- " + suggestion)
							}
						}
					} else if pegErr, ok := err.(*peg.Error); ok {
						fmt.Println(pegErr.Error())
					} else {
						fmt.Println(err)
					}
				}
			} else {
				fmt.Println("Parsing completed successfully")
			}

			if val != nil && (*astFlag || *optFlag) {
				ast := val.(*peg.Ast)
				if *optFlag {
					opt := peg.NewAstOptimizer(nil)
					ast = opt.Optimize(ast, nil)
				}
				fmt.Println(ast)
			}
		} else {
			// Use normal parsing
			val, err := parser.ParseAndGetValue(source, nil)
			pcheck(err)

			if *astFlag || *optFlag {
				ast := val.(*peg.Ast)
				if *optFlag {
					opt := peg.NewAstOptimizer(nil)
					ast = opt.Optimize(ast, nil)
				}
				fmt.Println(ast)
			}
		}
	}
}
