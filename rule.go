package peg

import (
	"fmt"
	"strings"
)

// Error detail
type ErrorDetail struct {
	Ln   int
	Col  int
	Msg  string
	Line string
}

func (d ErrorDetail) String() string {
	// Create the error message with better visualization
	str := fmt.Sprintf("Error at line %d, column %d: %s\n", d.Ln, d.Col, d.Msg)

	// Add the line of code where the error occurred
	if len(d.Line) > 0 {
		str += d.Line + "\n"

		// Create pointer to the exact error position
		pointer := strings.Repeat(" ", d.Col-1) + "^"
		str += pointer
	}

	return str
}

// Error types
type ErrorType int

const (
	SyntaxErrorType ErrorType = iota
	GrammarErrorType
	SemanticErrorType
)

// Error
type Error struct {
	Details []ErrorDetail
	Type    ErrorType
}

func (e *Error) Error() string {
	d := e.Details[0]
	return d.String()
}

// GetSuggestions provides helpful suggestions based on error context
func (e *Error) GetSuggestions() []string {
	var suggestions []string

	for _, detail := range e.Details {
		// Check for common error patterns and provide suggestions
		if strings.Contains(detail.Msg, "expected") {
			// For expected token errors, suggest possible corrections
			suggestions = append(suggestions, "Check if you have the correct syntax for this expression")
		} else if strings.Contains(detail.Msg, "not exact match") {
			suggestions = append(suggestions, "The input was partially parsed but did not match the entire grammar")
		}
	}

	return suggestions
}

// Create more specific error types
type SyntaxError struct {
	BaseError Error
	Expected  []string
}

// Implement the error interface for SyntaxError
func (e *SyntaxError) Error() string {
	return e.BaseError.Error()
}

// GetSuggestions provides helpful suggestions based on error context for SyntaxError
func (e *SyntaxError) GetSuggestions() []string {
	var suggestions []string

	// Add suggestions based on expected tokens
	if len(e.Expected) > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Expected one of: %s", strings.Join(e.Expected, ", ")))
	}

	// Add general suggestions
	suggestions = append(suggestions, "Check if you have the correct syntax for this expression")
	suggestions = append(suggestions, "Verify that your grammar rules are correctly defined")

	return suggestions
}

type GrammarError struct {
	BaseError Error
	RuleName  string
	ErrorType string // e.g., "left recursion", "undefined rule", etc.
}

// Implement the error interface for GrammarError
func (e *GrammarError) Error() string {
	return e.BaseError.Error()
}

// Action
type Action func(v *Values, d Any) (Any, error)

// Rule
type Rule struct {
	Name          string
	SS            string
	Pos           int
	Ope           operator
	Action        Action
	Enter         func(d Any)
	Leave         func(d Any)
	Message       func() (message string)
	Ignore        bool
	WhitespaceOpe operator
	WordOpe       operator

	Parameters []string

	TracerEnter func(name string, s string, v *Values, d Any, p int)
	TracerLeave func(name string, s string, v *Values, d Any, p int, l int)

	tokenChecker  *tokenChecker
	disableAction bool
}

func (r *Rule) Parse(s string, d Any) (l int, val Any, err error) {
	v := &Values{}
	c := &context{
		s:             s,
		errorPos:      -1,
		messagePos:    -1,
		whitespaceOpe: r.WhitespaceOpe,
		wordOpe:       r.WordOpe,
		tracerEnter:   r.TracerEnter,
		tracerLeave:   r.TracerLeave,
	}

	var ope operator = r
	if r.WhitespaceOpe != nil {
		ope = Seq(r.WhitespaceOpe, r) // Skip whitespace at beginning
	}

	l = ope.parse(s, 0, v, c, d)

	if success(l) && len(v.Vs) > 0 && v.Vs[0] != nil {
		val = v.Vs[0]
	}

	if fail(l) || l != len(s) {
		var pos int
		var msg string
		var line string
		if fail(l) {
			if c.messagePos > -1 {
				pos = c.messagePos
				msg = c.message
			} else {
				pos = c.errorPos
				ln, _ := lineInfo(s, pos)
				lineStart, lineEnd := printLine(s, ln)
				line = s[lineStart:lineEnd]

				// Enhanced error message with expected tokens
				if len(c.expectedTokens) > 0 {
					msg = fmt.Sprintf("Syntax error: expected %s", strings.Join(c.expectedTokens, ", "))
				} else {
					msg = "Syntax error"
				}
			}
		} else {
			msg = "not exact match"
			pos = l
		}
		ln, col := lineInfo(s, pos)

		// Create appropriate error type
		syntaxErr := &Error{
			Type: SyntaxErrorType,
		}

		if strings.Contains(msg, "expected") {
			syntaxErr.Details = append(syntaxErr.Details, ErrorDetail{ln, col, msg, line})
			err = &SyntaxError{
				BaseError: *syntaxErr,
				Expected:  c.expectedTokens,
			}
		} else {
			syntaxErr.Details = append(syntaxErr.Details, ErrorDetail{ln, col, msg, line})
			err = syntaxErr
		}

		// Details are already added to the error
	}

	return
}

func printLine(s string, line int) (int, int) {
	currentLine := 1
	currentOffset := 0
	lineLength := 0

	for currentLine < line && currentOffset < len(s) {
		if s[currentOffset] == '\n' {
			currentLine++
		}
		currentOffset++
	}

	for currentOffset+lineLength < len(s) && s[currentOffset+lineLength] != '\n' {
		lineLength++
	}

	return currentOffset, currentOffset + lineLength
}

func (o *Rule) Label() string {
	return fmt.Sprintf("[%s]", o.Name)
}

func (o *Rule) parse(s string, p int, v *Values, c *context, d Any) int {
	return parse(o, s, p, v, c, d)
}

func (r *Rule) parseCore(s string, p int, v *Values, c *context, d Any) int {
	// Macro reference
	if r.Parameters != nil {
		return r.Ope.parse(s, p, v, c, d)
	}

	if r.Enter != nil {
		r.Enter(d)
	}

	chv := c.push()

	l := r.Ope.parse(s, p, chv, c, d)

	// Invoke action
	var val Any

	if success(l) {
		if r.Action != nil && !r.disableAction {
			chv.S = s[p : p+l]
			chv.Pos = p

			var err error
			if val, err = r.Action(chv, d); err != nil {
				if c.messagePos < p {
					c.messagePos = p
					c.message = err.Error()
				}
				l = -1
			}
		} else if len(chv.Vs) > 0 {
			val = chv.Vs[0]
		}
	}

	if success(l) {
		if r.Ignore == false {
			v.Vs = append(v.Vs, val)
		}
	} else {
		if r.Message != nil {
			if c.messagePos < p {
				c.messagePos = p
				c.message = r.Message()
			}
		}
	}

	c.pop()

	if r.Leave != nil {
		r.Leave(d)
	}

	return l
}

func (r *Rule) accept(v visitor) {
	v.visitRule(r)
}

func (r *Rule) isToken() bool {
	if r.tokenChecker == nil {
		r.tokenChecker = &tokenChecker{}
		r.Ope.accept(r.tokenChecker)
	}
	return r.tokenChecker.isToken()
}

func (r *Rule) hasTokenBoundary() bool {
	if r.tokenChecker == nil {
		r.tokenChecker = &tokenChecker{}
		r.Ope.accept(r.tokenChecker)
	}
	return r.tokenChecker.hasTokenBoundary
}

// lineInfo
func lineInfo(s string, curPos int) (ln int, col int) {
	pos := 0
	colStartPos := 0
	ln = 1

	for pos < curPos {
		if s[pos] == '\n' {
			ln++
			colStartPos = pos + 1
		}
		pos++
	}

	col = pos - colStartPos + 1
	return
}
