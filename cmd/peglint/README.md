peglint
-------

The lint utility for PEG with enhanced error reporting and recovery.

```
usage: peglint [-ast] [-opt] [-trace] [-recovery] [-max-errors N] [-f path] [-s string] [grammar path]
```

peglint checks syntax of a given PEG grammar file and reports errors. If the check is successful and a user gives a source file for the grammar, it will also check syntax of the source file.

### Basic Options

The -ast flag prints the AST (abstract syntax tree) of the source file.

The -opt flag prints the optimized AST (abstract syntax tree) of the source file.

The -f 'path' specifies a file path to the source text.

The -s 'string' specifies the source text.

### Error Reporting and Recovery

peglint now provides enhanced error reporting with detailed context and suggestions:

- Line and column information
- Visual indicators pointing to the exact error location
- Expected tokens information
- Helpful suggestions for fixing errors

The -recovery flag enables error recovery mode, which allows parsing to continue after encountering errors.

The -max-errors N flag specifies the maximum number of errors to report before stopping (default: 10).

### Debugging

The -trace flag enables detailed tracing for debugging. It shows:

- Rule entry and exit events
- Success or failure of rule matching
- Error context when errors occur

### Examples

Basic usage:
```bash
peglint grammar.peg -f source.txt
```

With error recovery:
```bash
peglint -recovery -max-errors 5 grammar.peg -f source.txt
```

With tracing:
```bash
peglint -trace grammar.peg -f source.txt
```

With AST generation:
```bash
peglint -ast grammar.peg -f source.txt
```
