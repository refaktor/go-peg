package peg

import (
	"fmt"
	"strings"
)

const (
	WhitespceRuleName = "%whitespace"
	WordRuleName      = "%word"
	OptExpressionRule = "%expr"
	OptBinaryOperator = "%binop"
)

// PEG parser generator
type duplicate struct {
	name string
	pos  int
}

type data struct {
	grammar    map[string]*Rule
	start      string
	duplicates []duplicate
	options    map[string][]string
}

func newData() *data {
	return &data{
		grammar: make(map[string]*Rule),
		options: make(map[string][]string),
	}
}

var rStart, rDefinition, rExpression,
	rSequence, rPrefix, rSuffix, rPrimary,
	rIdentifier, rIdentCont, rIdentStart, rIdentRest,
	rLiteral, rClass, rRange, rChar,
	rLEFTARROW, rSLASH, rAND, rNOT, rQUESTION, rSTAR, rPLUS, rOPEN, rCLOSE, rDOT,
	rSpacing, rComment, rSpace, rEndOfLine, rEndOfFile, rBeginTok, rEndTok,
	rIgnore, rIGNORE,
	rParameters, rArguments, rCOMMA,
	rOption, rOptionValue, rOptionComment, rASSIGN, rSEPARATOR Rule

func init() {
	// Setup PEG syntax parser
	rStart.Ope = Seq(
		&rSpacing,
		Oom(&rDefinition),
		Opt(Seq(&rSEPARATOR, Oom(&rOption))),
		&rEndOfFile)

	rDefinition.Ope = Cho(
		Seq(&rIgnore, &rIdentCont, &rParameters, &rLEFTARROW, &rExpression),
		Seq(&rIgnore, &rIdentifier, &rLEFTARROW, &rExpression))

	rExpression.Ope = Seq(&rSequence, Zom(Seq(&rSLASH, &rSequence)))
	rSequence.Ope = Zom(&rPrefix)
	rPrefix.Ope = Seq(Opt(Cho(&rAND, &rNOT)), &rSuffix)
	rSuffix.Ope = Seq(&rPrimary, Opt(Cho(&rQUESTION, &rSTAR, &rPLUS)))

	rPrimary.Ope = Cho(
		Seq(&rIgnore, &rIdentCont, &rArguments, Npd(&rLEFTARROW)),
		Seq(&rIgnore, &rIdentifier, Npd(Seq(Opt(&rParameters), &rLEFTARROW))),
		Seq(&rOPEN, &rExpression, &rCLOSE),
		Seq(&rBeginTok, &rExpression, &rEndTok),
		&rLiteral,
		&rClass,
		&rDOT)

	rIdentifier.Ope = Seq(&rIdentCont, &rSpacing)
	rIdentCont.Ope = Seq(&rIdentStart, Zom(&rIdentRest))
	rIdentStart.Ope = Cls("a-zA-Z_\x80-\xff%")
	rIdentRest.Ope = Cho(&rIdentStart, Cls("0-9"))

	rLiteral.Ope = Cho(
		Seq(Lit("'"), Tok(Zom(Seq(Npd(Lit("'")), &rChar))), Lit("'"), &rSpacing),
		Seq(Lit("\""), Tok(Zom(Seq(Npd(Lit("\"")), &rChar))), Lit("\""), &rSpacing))

	rClass.Ope = Seq(Lit("["), Tok(Zom(Seq(Npd(Lit("]")), &rRange))), Lit("]"), &rSpacing)

	rRange.Ope = Cho(Seq(&rChar, Lit("-"), &rChar), &rChar)
	rChar.Ope = Cho(
		Seq(Lit("\\"), Cls("nrtfv'\"[]\\")),
		Seq(Lit("\\"), Cls("0-3"), Cls("0-7"), Cls("0-7")),
		Seq(Lit("\\"), Cls("0-7"), Opt(Cls("0-7"))),
		Seq(Lit("\\x"), Cls("0-9a-fA-F"), Opt(Cls("0-9a-fA-F"))),
		Seq(Npd(Lit("\\")), Dot()))

	rLEFTARROW.Ope = Seq(Cho(Lit("<-"), Lit("←")), &rSpacing)
	rSLASH.Ope = Seq(Lit("/"), &rSpacing)
	rSLASH.Ignore = true
	rAND.Ope = Seq(Lit("&"), &rSpacing)
	rNOT.Ope = Seq(Lit("!"), &rSpacing)
	rQUESTION.Ope = Seq(Lit("?"), &rSpacing)
	rSTAR.Ope = Seq(Lit("*"), &rSpacing)
	rPLUS.Ope = Seq(Lit("+"), &rSpacing)
	rOPEN.Ope = Seq(Lit("("), &rSpacing)
	rOPEN.Ignore = true
	rCLOSE.Ope = Seq(Lit(")"), &rSpacing)
	rCLOSE.Ignore = true
	rDOT.Ope = Seq(Lit("."), &rSpacing)

	rSpacing.Ope = Zom(Cho(&rSpace, &rComment))
	rComment.Ope = Seq(Lit("#"), Zom(Seq(Npd(&rEndOfLine), Dot())), &rEndOfLine)
	rSpace.Ope = Cho(Lit(" "), Lit("\t"), &rEndOfLine)
	rEndOfLine.Ope = Cho(Lit("\r\n"), Lit("\n"), Lit("\r"))
	rEndOfFile.Ope = Npd(Dot())

	rBeginTok.Ope = Seq(Lit("<"), &rSpacing)
	rBeginTok.Ignore = true
	rEndTok.Ope = Seq(Lit(">"), &rSpacing)
	rEndTok.Ignore = true

	rIGNORE.Ope = Lit("~")
	rSEPARATOR.Ope = Seq(Lit("---"), &rSpacing)

	rIgnore.Ope = Opt(&rIGNORE)

	rParameters.Ope = Seq(&rOPEN, &rIdentifier, Zom(Seq(&rCOMMA, &rIdentifier)), &rCLOSE)
	rArguments.Ope = Seq(&rOPEN, &rExpression, Zom(Seq(&rCOMMA, &rExpression)), &rCLOSE)
	rCOMMA.Ope = Seq(Lit(","), &rSpacing)
	rCOMMA.Ignore = true

	rOption.Ope = Seq(&rIdentifier, &rASSIGN, &rOptionValue)
	rOptionComment.Ope = Seq(Zom(Cho(Lit(" "), Lit("\t"))), Cho(&rComment, &rEndOfLine))
	rOptionValue.Ope = Seq(Tok(Zom(Seq(Npd(&rOptionComment), Dot()))), &rOptionComment, &rSpacing)
	rASSIGN.Ope = Seq(Lit("="), &rSpacing)
	rSEPARATOR.Ope = Seq(Lit("---"), &rSpacing)

	// Setup actions
	rDefinition.Action = func(v *Values, d Any) (val Any, err error) {
		var ignore bool
		var name string
		var params []string
		var ope operator

		switch v.Choice {
		case 0: // Macro
			ignore = v.ToBool(0)
			name = v.ToStr(1)
			params = v.Vs[2].([]string)
			ope = v.ToOpe(4)
		case 1: // Rule
			ignore = v.ToBool(0)
			name = v.ToStr(1)
			ope = v.ToOpe(3)
		}

		data := d.(*data)
		_, ok := data.grammar[name]
		if ok {
			data.duplicates = append(data.duplicates, duplicate{name, v.Pos})
		} else {
			data.grammar[name] = &Rule{
				Ope:        ope,
				Name:       name,
				SS:         v.SS,
				Pos:        v.Pos,
				Ignore:     ignore,
				Parameters: params,
			}
			if len(data.start) == 0 {
				data.start = name
			}
		}
		return
	}

	rParameters.Action = func(v *Values, d Any) (val Any, err error) {
		var params []string
		for i := 0; i < len(v.Vs); i++ {
			params = append(params, v.ToStr(i))
		}
		val = params
		return
	}

	rArguments.Action = func(v *Values, d Any) (val Any, err error) {
		var exprs []operator
		for i := 0; i < len(v.Vs); i++ {
			exprs = append(exprs, v.ToOpe(i))
		}
		val = exprs
		return
	}

	rExpression.Action = func(v *Values, d Any) (val Any, err error) {
		if len(v.Vs) == 1 {
			val = v.ToOpe(0)
		} else {
			var opes []operator
			for i := 0; i < len(v.Vs); i++ {
				opes = append(opes, v.ToOpe(i))
			}
			val = Cho(opes...)
		}
		return
	}

	rSequence.Action = func(v *Values, d Any) (val Any, err error) {
		if len(v.Vs) == 1 {
			val = v.ToOpe(0)
		} else {
			var opes []operator
			for i := 0; i < len(v.Vs); i++ {
				opes = append(opes, v.ToOpe(i))
			}
			val = Seq(opes...)
		}
		return
	}

	rPrefix.Action = func(v *Values, d Any) (val Any, err error) {
		if len(v.Vs) == 1 {
			val = v.ToOpe(0)
		} else {
			tok := v.ToStr(0)
			ope := v.ToOpe(1)
			switch tok {
			case "&":
				val = Apd(ope)
			case "!":
				val = Npd(ope)
			}
		}
		return
	}

	rSuffix.Action = func(v *Values, d Any) (val Any, err error) {
		ope := v.ToOpe(0)
		if len(v.Vs) == 1 {
			val = ope
		} else {
			tok := v.ToStr(1)
			switch tok {
			case "?":
				val = Opt(ope)
			case "*":
				val = Zom(ope)
			case "+":
				val = Oom(ope)
			}
		}
		return
	}

	rPrimary.Action = func(v *Values, d Any) (val Any, err error) {
		switch v.Choice {
		case 0 /* Macro Reference */, 1: /* Reference */
			ignore := v.ToBool(0)
			ident := v.ToStr(1)

			var args []operator
			if v.Choice == 0 {
				args = v.Vs[2].([]operator)
			}

			if ignore {
				val = Ign(Ref(ident, args, v.Pos))
			} else {
				val = Ref(ident, args, v.Pos)
			}
		case 2: // Expression
			val = v.ToOpe(0)
		case 3: // TokenBoundary
			val = Tok(v.ToOpe(0))
		default:
			val = v.ToOpe(0)
		}
		return
	}

	rIdentCont.Action = func(v *Values, d Any) (Any, error) {
		return v.S, nil
	}

	rLiteral.Action = func(v *Values, d Any) (Any, error) {
		return Lit(resolveEscapeSequence(v.Ts[0].S)), nil
	}

	rClass.Action = func(v *Values, d Any) (Any, error) {
		return Cls(resolveEscapeSequence(v.Ts[0].S)), nil
	}

	rAND.Action = func(v *Values, d Any) (Any, error) {
		return v.S[:1], nil
	}
	rNOT.Action = func(v *Values, d Any) (Any, error) {
		return v.S[:1], nil
	}
	rQUESTION.Action = func(v *Values, d Any) (Any, error) {
		return v.S[:1], nil
	}
	rSTAR.Action = func(v *Values, d Any) (Any, error) {
		return v.S[:1], nil
	}
	rPLUS.Action = func(v *Values, d Any) (Any, error) {
		return v.S[:1], nil
	}

	rDOT.Action = func(v *Values, d Any) (Any, error) {
		return Dot(), nil
	}

	rIgnore.Action = func(v *Values, d Any) (val Any, err error) {
		val = len(v.Vs) != 0
		return
	}

	rOption.Action = func(v *Values, d Any) (val Any, err error) {
		options := d.(*data).options
		optName := v.ToStr(0)
		optVal := v.ToStr(2)
		options[optName] = append(options[optName], optVal)
		return
	}
	rOptionValue.Action = func(v *Values, d Any) (Any, error) {
		return v.Token(), nil
	}
}

func isHex(c byte) (v int, ok bool) {
	if '0' <= c && c <= '9' {
		v = int(c - '0')
		ok = true
	} else if 'a' <= c && c <= 'f' {
		v = int(c - 'a' + 10)
		ok = true
	} else if 'A' <= c && c <= 'F' {
		v = int(c - 'A' + 10)
		ok = true
	}
	return
}

func isDigit(c byte) (v int, ok bool) {
	if '0' <= c && c <= '9' {
		v = int(c - '0')
		ok = true
	}
	return
}

func parseHexNumber(s string, i int) (byte, int) {
	ret := 0
	for i < len(s) {
		val, ok := isHex(s[i])
		if !ok {
			break
		}
		ret = ret*16 + val
		i++
	}
	return byte(ret), i
}

func parseOctNumber(s string, i int) (byte, int) {
	ret := 0
	for i < len(s) {
		val, ok := isDigit(s[i])
		if !ok {
			break
		}
		ret = ret*8 + val
		i++
	}
	return byte(ret), i
}

func resolveEscapeSequence(s string) string {
	n := len(s)
	b := make([]byte, 0, n)

	i := 0
	for i < n {
		ch := s[i]
		if ch == '\\' {
			i++
			switch s[i] {
			case 'n':
				b = append(b, '\n')
				i++
			case 'r':
				b = append(b, '\r')
				i++
			case 't':
				b = append(b, '\t')
				i++
			case 'f':
				b = append(b, '\f')
				i++
			case 'v':
				b = append(b, '\v')
				i++
			case '\'':
				b = append(b, '\'')
				i++
			case '"':
				b = append(b, '"')
				i++
			case '[':
				b = append(b, '[')
				i++
			case ']':
				b = append(b, ']')
				i++
			case '\\':
				b = append(b, '\\')
				i++
			case 'x':
				ch, i = parseHexNumber(s, i+1)
				b = append(b, ch)
			default:
				ch, i = parseOctNumber(s, i)
				b = append(b, ch)
			}
		} else {
			b = append(b, ch)
			i++
		}
	}

	return string(b)
}

func getExpressionParsingOptions(options map[string][]string) (name string, info BinOpeInfo) {
	name = ""
	if vs, ok := options[OptExpressionRule]; ok {
		name = vs[0]
		// TODO: error handling
	}

	info = make(BinOpeInfo)
	if vs, ok := options[OptBinaryOperator]; ok {
		level := 1
		for _, s := range vs {
			flds := strings.Split(s, " ")
			// TODO: error handling
			assoc := assocNone
			for i, fld := range flds {
				switch i {
				case 0:
					switch fld {
					case "L":
						assoc = assocLeft
					case "R":
						assoc = assocRight
					default:
						// TODO: error handling
					}
				default:
					info[fld] = struct {
						level int
						assoc int
					}{level, assoc}
				}
			}
			level++
		}
	}

	return
}

// TracingOptions defines the configuration for parser tracing
type TracingOptions struct {
	ShowRuleEntry    bool   // Show when entering a rule
	ShowRuleExit     bool   // Show when exiting a rule
	ShowTokens       bool   // Show tokens as they are parsed
	ShowErrorContext bool   // Show context around errors
	OutputFormat     string // Format for output: "text" or "json"
}

// Parser
type Parser struct {
	Grammar         map[string]*Rule
	start           string
	TracerEnter     func(name string, s string, v *Values, d Any, p int)
	TracerLeave     func(name string, s string, v *Values, d Any, p int, l int)
	RecoveryEnabled bool            // Enable error recovery
	MaxErrors       int             // Maximum number of errors to report before stopping
	TracingOptions  *TracingOptions // Options for tracing
}

// findNextMeaningfulToken attempts to find the next token to continue parsing after an error
func findNextMeaningfulToken(s string, pos int, delimiters []string) int {
	if pos >= len(s) {
		return pos
	}

	// Skip whitespace
	for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t' || s[pos] == '\n' || s[pos] == '\r') {
		pos++
	}

	// If we reached the end, return
	if pos >= len(s) {
		return pos
	}

	// Check for delimiters
	for _, delimiter := range delimiters {
		if pos+len(delimiter) <= len(s) && s[pos:pos+len(delimiter)] == delimiter {
			return pos + len(delimiter)
		}
	}

	// If no delimiter found, skip to the next whitespace or end
	start := pos
	for pos < len(s) && s[pos] != ' ' && s[pos] != '\t' && s[pos] != '\n' && s[pos] != '\r' {
		pos++
	}

	// If we didn't move, just advance one character to avoid infinite loops
	if pos == start {
		pos++
	}

	return pos
}

// EnableTracing sets up tracing with the specified options
func (p *Parser) EnableTracing(options *TracingOptions) {
	p.TracingOptions = options

	if options == nil {
		// Default options if none provided
		options = &TracingOptions{
			ShowRuleEntry:    true,
			ShowRuleExit:     true,
			ShowTokens:       false,
			ShowErrorContext: true,
			OutputFormat:     "text",
		}
	}

	// Set up tracers based on options
	if options.ShowRuleEntry || options.ShowRuleExit {
		indent := func(level int) string {
			s := ""
			for level > 0 {
				s = s + "  "
				level--
			}
			return s
		}

		level := 0
		prevPos := 0

		if options.ShowRuleEntry {
			p.TracerEnter = func(name string, s string, v *Values, d Any, p int) {
				var backtrack string
				if p < prevPos {
					backtrack = "*"
				}
				fmt.Printf("%d:%d%s\t%s%s\n", p, level, backtrack, indent(level), name)
				prevPos = p
				level++
			}
		}

		if options.ShowRuleExit {
			p.TracerLeave = func(name string, s string, v *Values, d Any, p int, l int) {
				level--
				if l >= 0 {
					fmt.Printf("%d:%d\t%s%s (SUCCESS, len=%d)\n", p, level, indent(level), name, l)
				} else {
					fmt.Printf("%d:%d\t%s%s (FAILED)\n", p, level, indent(level), name)
				}
			}
		}
	}
}

func NewParser(s string) (p *Parser, err error) {
	return NewParserWithUserRules(s, nil)
}

func NewParserWithUserRules(s string, rules map[string]operator) (p *Parser, err error) {
	data := newData()

	_, _, err = rStart.Parse(s, data)
	if err != nil {
		return nil, err
	}

	// User provided rules
	for name, ope := range rules {
		ignore := false

		if len(name) > 0 && name[0] == '~' {
			ignore = true
			name = name[1:]
		}

		if len(name) > 0 {
			data.grammar[name] = &Rule{
				Ope:    ope,
				Name:   name,
				Ignore: ignore,
			}
		}
	}

	// Check duplicated definitions
	if len(data.duplicates) > 0 {
		err = &Error{}
		for _, dup := range data.duplicates {
			ln, col := lineInfo(s, dup.pos)
			msg := "'" + dup.name + "' is already defined."
			err.(*Error).Details = append(err.(*Error).Details, ErrorDetail{ln, col, msg, ""})
		}
	}

	// Check missing definitions
	for _, r := range data.grammar {
		v := &referenceChecker{
			grammar:  data.grammar,
			params:   r.Parameters,
			errorPos: make(map[string]int),
			errorMsg: make(map[string]string),
		}
		r.accept(v)
		for name, pos := range v.errorPos {
			if err == nil {
				err = &Error{}
			}
			ln, col := lineInfo(s, pos)
			msg := v.errorMsg[name]
			err.(*Error).Details = append(err.(*Error).Details, ErrorDetail{ln, col, msg, ""})
		}
	}

	if err != nil {
		return nil, err
	}

	// Link references
	for _, r := range data.grammar {
		v := &linkReferences{
			parameters: r.Parameters,
			grammar:    data.grammar,
		}
		r.accept(v)
	}

	// Check left recursion
	for name, r := range data.grammar {
		v := &detectLeftRecursion{
			pos:    -1,
			name:   name,
			params: r.Parameters,
			refs:   make(map[string]bool),
			done:   false,
		}
		r.accept(v)
		if v.pos != -1 {
			if err == nil {
				err = &Error{}
			}
			ln, col := lineInfo(s, v.pos)
			msg := "'" + name + "' is left recursive."
			err.(*Error).Details = append(err.(*Error).Details, ErrorDetail{ln, col, msg, ""})
		}
	}

	if err != nil {
		return nil, err
	}

	// Automatic whitespace skipping
	if r, ok := data.grammar[WhitespceRuleName]; ok {
		data.grammar[data.start].WhitespaceOpe = Wsp(r)
	}

	// Word expression
	if r, ok := data.grammar[WordRuleName]; ok {
		data.grammar[data.start].WordOpe = r
	}

	p = &Parser{
		Grammar: data.grammar,
		start:   data.start,
	}

	// Setup expression parsing
	name, info := getExpressionParsingOptions(data.options)
	err = EnableExpressionParsing(p, name, info)

	return
}

func (p *Parser) Parse(s string, d Any) (err error) {
	_, err = p.ParseAndGetValue(s, d)
	return
}

// ParseWithRecovery parses the input string with error recovery
func (p *Parser) ParseWithRecovery(s string, d Any) (errs []error) {
	if !p.RecoveryEnabled {
		err := p.Parse(s, d)
		if err != nil {
			errs = append(errs, err)
		}
		return
	}

	// Set default max errors if not specified
	maxErrors := p.MaxErrors
	if maxErrors <= 0 {
		maxErrors = 10 // Default to 10 errors
	}

	pos := 0
	for pos < len(s) {
		// Try to parse from current position
		r := p.Grammar[p.start]
		r.TracerEnter = p.TracerEnter
		r.TracerLeave = p.TracerLeave

		l, _, err := r.Parse(s[pos:], d)

		if err == nil {
			// Successful parse
			pos += l
		} else {
			// Error occurred
			errs = append(errs, err)

			if len(errs) >= maxErrors {
				break
			}

			// Try to recover by finding next meaningful token
			// Common delimiters in PEG grammars
			delimiters := []string{";", "}", "{", "(", ")", ",", "=", "<-"}
			newPos := findNextMeaningfulToken(s, pos+1, delimiters)

			// If we couldn't advance, just move one character to avoid infinite loops
			if newPos <= pos {
				newPos = pos + 1
			}

			pos = newPos
		}
	}

	return
}

func (p *Parser) ParseAndGetValue(s string, d Any) (val Any, err error) {
	r := p.Grammar[p.start]
	r.TracerEnter = p.TracerEnter
	r.TracerLeave = p.TracerLeave
	_, val, err = r.Parse(s, d)

	// Show error context if enabled
	if err != nil && p.TracingOptions != nil && p.TracingOptions.ShowErrorContext {
		fmt.Println("\nError Context:")
		fmt.Println(err.Error())

		// If it's a syntax error with expected tokens, show suggestions
		if syntaxErr, ok := err.(*SyntaxError); ok {
			fmt.Println("\nSuggestions:")
			for _, suggestion := range syntaxErr.GetSuggestions() {
				fmt.Println("- " + suggestion)
			}
		}
	}

	return
}

// ParseAndGetValueWithRecovery parses the input string with error recovery and returns the value
func (p *Parser) ParseAndGetValueWithRecovery(s string, d Any) (val Any, errs []error) {
	if !p.RecoveryEnabled {
		var err error
		val, err = p.ParseAndGetValue(s, d)
		if err != nil {
			errs = append(errs, err)
		}
		return
	}

	// Set default max errors if not specified
	maxErrors := p.MaxErrors
	if maxErrors <= 0 {
		maxErrors = 10 // Default to 10 errors
	}

	pos := 0
	for pos < len(s) {
		// Try to parse from current position
		r := p.Grammar[p.start]
		r.TracerEnter = p.TracerEnter
		r.TracerLeave = p.TracerLeave

		l, v, err := r.Parse(s[pos:], d)

		if err == nil {
			// Successful parse
			pos += l
			val = v // Use the last successful value
		} else {
			// Error occurred
			errs = append(errs, err)

			if len(errs) >= maxErrors {
				break
			}

			// Try to recover by finding next meaningful token
			// Common delimiters in PEG grammars
			delimiters := []string{";", "}", "{", "(", ")", ",", "=", "<-"}
			newPos := findNextMeaningfulToken(s, pos+1, delimiters)

			// If we couldn't advance, just move one character to avoid infinite loops
			if newPos <= pos {
				newPos = pos + 1
			}

			pos = newPos
		}
	}

	return
}
