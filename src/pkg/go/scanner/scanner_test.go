// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scanner

import (
	"go/token"
	"os"
	"testing"
)


const /* class */ (
	special = iota
	literal
	operator
	keyword
)


func tokenclass(tok token.Token) int {
	switch {
	case tok.IsLiteral():
		return literal
	case tok.IsOperator():
		return operator
	case tok.IsKeyword():
		return keyword
	}
	return special
}


type elt struct {
	tok   token.Token
	lit   string
	class int
}


var tokens = [...]elt{
	// Special tokens
	{token.COMMENT, "/* a comment */", special},
	{token.COMMENT, "// a comment \n", special},

	// Identifiers and basic type literals
	{token.IDENT, "foobar", literal},
	{token.IDENT, "a۰۱۸", literal},
	{token.IDENT, "foo६४", literal},
	{token.IDENT, "bar９８７６", literal},
	{token.INT, "0", literal},
	{token.INT, "1", literal},
	{token.INT, "123456789012345678890", literal},
	{token.INT, "01234567", literal},
	{token.INT, "0xcafebabe", literal},
	{token.FLOAT, "0.", literal},
	{token.FLOAT, ".0", literal},
	{token.FLOAT, "3.14159265", literal},
	{token.FLOAT, "1e0", literal},
	{token.FLOAT, "1e+100", literal},
	{token.FLOAT, "1e-100", literal},
	{token.FLOAT, "2.71828e-1000", literal},
	{token.IMAG, "0i", literal},
	{token.IMAG, "1i", literal},
	{token.IMAG, "012345678901234567889i", literal},
	{token.IMAG, "123456789012345678890i", literal},
	{token.IMAG, "0.i", literal},
	{token.IMAG, ".0i", literal},
	{token.IMAG, "3.14159265i", literal},
	{token.IMAG, "1e0i", literal},
	{token.IMAG, "1e+100i", literal},
	{token.IMAG, "1e-100i", literal},
	{token.IMAG, "2.71828e-1000i", literal},
	{token.CHAR, "'a'", literal},
	{token.CHAR, "'\\000'", literal},
	{token.CHAR, "'\\xFF'", literal},
	{token.CHAR, "'\\uff16'", literal},
	{token.CHAR, "'\\U0000ff16'", literal},
	{token.STRING, "`foobar`", literal},
	{token.STRING, "`" + `foo
	                        bar` +
		"`",
		literal,
	},

	// Operators and delimitors
	{token.ADD, "+", operator},
	{token.SUB, "-", operator},
	{token.MUL, "*", operator},
	{token.QUO, "/", operator},
	{token.REM, "%", operator},

	{token.AND, "&", operator},
	{token.OR, "|", operator},
	{token.XOR, "^", operator},
	{token.SHL, "<<", operator},
	{token.SHR, ">>", operator},
	{token.AND_NOT, "&^", operator},

	{token.ADD_ASSIGN, "+=", operator},
	{token.SUB_ASSIGN, "-=", operator},
	{token.MUL_ASSIGN, "*=", operator},
	{token.QUO_ASSIGN, "/=", operator},
	{token.REM_ASSIGN, "%=", operator},

	{token.AND_ASSIGN, "&=", operator},
	{token.OR_ASSIGN, "|=", operator},
	{token.XOR_ASSIGN, "^=", operator},
	{token.SHL_ASSIGN, "<<=", operator},
	{token.SHR_ASSIGN, ">>=", operator},
	{token.AND_NOT_ASSIGN, "&^=", operator},

	{token.LAND, "&&", operator},
	{token.LOR, "||", operator},
	{token.ARROW, "<-", operator},
	{token.INC, "++", operator},
	{token.DEC, "--", operator},

	{token.EQL, "==", operator},
	{token.LSS, "<", operator},
	{token.GTR, ">", operator},
	{token.ASSIGN, "=", operator},
	{token.NOT, "!", operator},

	{token.NEQ, "!=", operator},
	{token.LEQ, "<=", operator},
	{token.GEQ, ">=", operator},
	{token.DEFINE, ":=", operator},
	{token.ELLIPSIS, "...", operator},

	{token.LPAREN, "(", operator},
	{token.LBRACK, "[", operator},
	{token.LBRACE, "{", operator},
	{token.COMMA, ",", operator},
	{token.PERIOD, ".", operator},

	{token.RPAREN, ")", operator},
	{token.RBRACK, "]", operator},
	{token.RBRACE, "}", operator},
	{token.SEMICOLON, ";", operator},
	{token.COLON, ":", operator},

	// Keywords
	{token.BREAK, "break", keyword},
	{token.CASE, "case", keyword},
	{token.CHAN, "chan", keyword},
	{token.CONST, "const", keyword},
	{token.CONTINUE, "continue", keyword},

	{token.DEFAULT, "default", keyword},
	{token.DEFER, "defer", keyword},
	{token.ELSE, "else", keyword},
	{token.FALLTHROUGH, "fallthrough", keyword},
	{token.FOR, "for", keyword},

	{token.FUNC, "func", keyword},
	{token.GO, "go", keyword},
	{token.GOTO, "goto", keyword},
	{token.IF, "if", keyword},
	{token.IMPORT, "import", keyword},

	{token.INTERFACE, "interface", keyword},
	{token.MAP, "map", keyword},
	{token.PACKAGE, "package", keyword},
	{token.RANGE, "range", keyword},
	{token.RETURN, "return", keyword},

	{token.SELECT, "select", keyword},
	{token.STRUCT, "struct", keyword},
	{token.SWITCH, "switch", keyword},
	{token.TYPE, "type", keyword},
	{token.VAR, "var", keyword},
}


const whitespace = "  \t  \n\n\n" // to separate tokens

type testErrorHandler struct {
	t *testing.T
}

func (h *testErrorHandler) Error(pos token.Position, msg string) {
	h.t.Errorf("Error() called (msg = %s)", msg)
}


func newlineCount(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			n++
		}
	}
	return n
}


func checkPos(t *testing.T, lit string, pos, expected token.Position) {
	if pos.Filename != expected.Filename {
		t.Errorf("bad filename for %s: got %s, expected %s", lit, pos.Filename, expected.Filename)
	}
	if pos.Offset != expected.Offset {
		t.Errorf("bad position for %s: got %d, expected %d", lit, pos.Offset, expected.Offset)
	}
	if pos.Line != expected.Line {
		t.Errorf("bad line for %s: got %d, expected %d", lit, pos.Line, expected.Line)
	}
	if pos.Column != expected.Column {
		t.Errorf("bad column for %s: got %d, expected %d", lit, pos.Column, expected.Column)
	}
}


// Verify that calling Scan() provides the correct results.
func TestScan(t *testing.T) {
	// make source
	var src string
	for _, e := range tokens {
		src += e.lit + whitespace
	}
	src_linecount := newlineCount(src)
	whitespace_linecount := newlineCount(whitespace)

	// verify scan
	index := 0
	epos := token.Position{"", 0, 1, 1} // expected position
	nerrors := Tokenize("", []byte(src), &testErrorHandler{t}, ScanComments,
		func(pos token.Position, tok token.Token, litb []byte) bool {
			e := elt{token.EOF, "", special}
			if index < len(tokens) {
				e = tokens[index]
			}
			lit := string(litb)
			if tok == token.EOF {
				lit = "<EOF>"
				epos.Line = src_linecount
				epos.Column = 1
			}
			checkPos(t, lit, pos, epos)
			if tok != e.tok {
				t.Errorf("bad token for %q: got %s, expected %s", lit, tok.String(), e.tok.String())
			}
			if e.tok.IsLiteral() && lit != e.lit {
				t.Errorf("bad literal for %q: got %q, expected %q", lit, lit, e.lit)
			}
			if tokenclass(tok) != e.class {
				t.Errorf("bad class for %q: got %d, expected %d", lit, tokenclass(tok), e.class)
			}
			epos.Offset += len(lit) + len(whitespace)
			epos.Line += newlineCount(lit) + whitespace_linecount
			if tok == token.COMMENT && litb[1] == '/' {
				// correct for unaccounted '/n' in //-style comment
				epos.Offset++
				epos.Line++
			}
			index++
			return tok != token.EOF
		})
	if nerrors != 0 {
		t.Errorf("found %d errors", nerrors)
	}
}


func checkSemi(t *testing.T, line string, mode uint) {
	var S Scanner
	S.Init("TestSemis", []byte(line), nil, mode)
	pos, tok, lit := S.Scan()
	for tok != token.EOF {
		if tok == token.ILLEGAL {
			// the illegal token literal indicates what
			// kind of semicolon literal to expect
			semiLit := "\n"
			if lit[0] == '#' {
				semiLit = ";"
			}
			// next token must be a semicolon
			offs := pos.Offset + 1
			pos, tok, lit = S.Scan()
			if tok == token.SEMICOLON {
				if pos.Offset != offs {
					t.Errorf("bad offset for %q: got %d, expected %d", line, pos.Offset, offs)
				}
				if string(lit) != semiLit {
					t.Errorf(`bad literal for %q: got %q, expected %q`, line, lit, semiLit)
				}
			} else {
				t.Errorf("bad token for %q: got %s, expected ;", line, tok.String())
			}
		} else if tok == token.SEMICOLON {
			t.Errorf("bad token for %q: got ;, expected no ;", line)
		}
		pos, tok, lit = S.Scan()
	}
}


var lines = []string{
	// # indicates a semicolon present in the source
	// $ indicates an automatically inserted semicolon
	"",
	"#;",
	"foo$\n",
	"123$\n",
	"1.2$\n",
	"'x'$\n",
	`"x"` + "$\n",
	"`x`$\n",

	"+\n",
	"-\n",
	"*\n",
	"/\n",
	"%\n",

	"&\n",
	"|\n",
	"^\n",
	"<<\n",
	">>\n",
	"&^\n",

	"+=\n",
	"-=\n",
	"*=\n",
	"/=\n",
	"%=\n",

	"&=\n",
	"|=\n",
	"^=\n",
	"<<=\n",
	">>=\n",
	"&^=\n",

	"&&\n",
	"||\n",
	"<-\n",
	"++$\n",
	"--$\n",

	"==\n",
	"<\n",
	">\n",
	"=\n",
	"!\n",

	"!=\n",
	"<=\n",
	">=\n",
	":=\n",
	"...\n",

	"(\n",
	"[\n",
	"{\n",
	",\n",
	".\n",

	")$\n",
	"]$\n",
	"}$\n",
	"#;\n",
	":\n",

	"break$\n",
	"case\n",
	"chan\n",
	"const\n",
	"continue$\n",

	"default\n",
	"defer\n",
	"else\n",
	"fallthrough$\n",
	"for\n",

	"func\n",
	"go\n",
	"goto\n",
	"if\n",
	"import\n",

	"interface\n",
	"map\n",
	"package\n",
	"range\n",
	"return$\n",

	"select\n",
	"struct\n",
	"switch\n",
	"type\n",
	"var\n",

	"foo$//comment\n",
	"foo$/*comment*/\n",
	"foo$/*\n*/",
	"foo$/*comment*/    \n",
	"foo$/*\n*/    ",
	"foo    $// comment\n",
	"foo    $/*comment*/\n",
	"foo    $/*\n*/",

	"foo    $/*0*/ /*1*/ /*2*/\n",
	"foo    $/*comment*/    \n",
	"foo    $/*0*/ /*1*/ /*2*/    \n",
	"foo	$/**/ /*-------------*/       /*----\n*/bar       $/*  \n*/baa$\n",

	"package main$\n\nfunc main() {\n\tif {\n\t\treturn /* */ }$\n}$\n",
}


func TestSemis(t *testing.T) {
	for _, line := range lines {
		checkSemi(t, line, AllowIllegalChars|InsertSemis)
		checkSemi(t, line, AllowIllegalChars|InsertSemis|ScanComments)

		// if the input ended in newlines, the input must tokenize the
		// same with or without those newlines
		for i := len(line) - 1; i >= 0 && line[i] == '\n'; i-- {
			checkSemi(t, line[0:i], AllowIllegalChars|InsertSemis)
			checkSemi(t, line[0:i], AllowIllegalChars|InsertSemis|ScanComments)
		}
	}
}


type seg struct {
	srcline  string // a line of source text
	filename string // filename for current token
	line     int    // line number for current token
}


var segments = []seg{
	// exactly one token per line since the test consumes one token per segment
	{"  line1", "TestLineComments", 1},
	{"\nline2", "TestLineComments", 2},
	{"\nline3  //line File1.go:100", "TestLineComments", 3}, // bad line comment, ignored
	{"\nline4", "TestLineComments", 4},
	{"\n//line File1.go:100\n  line100", "File1.go", 100},
	{"\n//line File2.go:200\n  line200", "File2.go", 200},
	{"\n//line :1\n  line1", "", 1},
	{"\n//line foo:42\n  line42", "foo", 42},
	{"\n //line foo:42\n  line44", "foo", 44},           // bad line comment, ignored
	{"\n//line foo 42\n  line46", "foo", 46},            // bad line comment, ignored
	{"\n//line foo:42 extra text\n  line48", "foo", 48}, // bad line comment, ignored
	{"\n//line foo:42\n  line42", "foo", 42},
	{"\n//line foo:42\n  line42", "foo", 42},
	{"\n//line File1.go:100\n  line100", "File1.go", 100},
}


// Verify that comments of the form "//line filename:line" are interpreted correctly.
func TestLineComments(t *testing.T) {
	// make source
	var src string
	for _, e := range segments {
		src += e.srcline
	}

	// verify scan
	var S Scanner
	S.Init("TestLineComments", []byte(src), nil, 0)
	for _, s := range segments {
		pos, _, lit := S.Scan()
		checkPos(t, string(lit), pos, token.Position{s.filename, pos.Offset, s.line, pos.Column})
	}

	if S.ErrorCount != 0 {
		t.Errorf("found %d errors", S.ErrorCount)
	}
}


// Verify that initializing the same scanner more then once works correctly.
func TestInit(t *testing.T) {
	var s Scanner

	// 1st init
	s.Init("", []byte("if true { }"), nil, 0)
	s.Scan()              // if
	s.Scan()              // true
	_, tok, _ := s.Scan() // {
	if tok != token.LBRACE {
		t.Errorf("bad token: got %s, expected %s", tok.String(), token.LBRACE)
	}

	// 2nd init
	s.Init("", []byte("go true { ]"), nil, 0)
	_, tok, _ = s.Scan() // go
	if tok != token.GO {
		t.Errorf("bad token: got %s, expected %s", tok.String(), token.GO)
	}

	if s.ErrorCount != 0 {
		t.Errorf("found %d errors", s.ErrorCount)
	}
}


func TestIllegalChars(t *testing.T) {
	var s Scanner

	const src = "*?*$*@*"
	s.Init("", []byte(src), &testErrorHandler{t}, AllowIllegalChars)
	for offs, ch := range src {
		pos, tok, lit := s.Scan()
		if pos.Offset != offs {
			t.Errorf("bad position for %s: got %d, expected %d", string(lit), pos.Offset, offs)
		}
		if tok == token.ILLEGAL && string(lit) != string(ch) {
			t.Errorf("bad token: got %s, expected %s", string(lit), string(ch))
		}
	}

	if s.ErrorCount != 0 {
		t.Errorf("found %d errors", s.ErrorCount)
	}
}


func TestStdErrorHander(t *testing.T) {
	const src = "@\n" + // illegal character, cause an error
		"@ @\n" + // two errors on the same line
		"//line File2:20\n" +
		"@\n" + // different file, but same line
		"//line File2:1\n" +
		"@ @\n" + // same file, decreasing line number
		"//line File1:1\n" +
		"@ @ @" // original file, line 1 again

	v := new(ErrorVector)
	nerrors := Tokenize("File1", []byte(src), v, 0,
		func(pos token.Position, tok token.Token, litb []byte) bool {
			return tok != token.EOF
		})

	list := v.GetErrorList(Raw)
	if len(list) != 9 {
		t.Errorf("found %d raw errors, expected 9", len(list))
		PrintError(os.Stderr, list)
	}

	list = v.GetErrorList(Sorted)
	if len(list) != 9 {
		t.Errorf("found %d sorted errors, expected 9", len(list))
		PrintError(os.Stderr, list)
	}

	list = v.GetErrorList(NoMultiples)
	if len(list) != 4 {
		t.Errorf("found %d one-per-line errors, expected 4", len(list))
		PrintError(os.Stderr, list)
	}

	if v.ErrorCount() != nerrors {
		t.Errorf("found %d errors, expected %d", v.ErrorCount(), nerrors)
	}
}


type errorCollector struct {
	cnt int            // number of errors encountered
	msg string         // last error message encountered
	pos token.Position // last error position encountered
}


func (h *errorCollector) Error(pos token.Position, msg string) {
	h.cnt++
	h.msg = msg
	h.pos = pos
}


func checkError(t *testing.T, src string, tok token.Token, pos int, err string) {
	var s Scanner
	var h errorCollector
	s.Init("", []byte(src), &h, ScanComments)
	_, tok0, _ := s.Scan()
	_, tok1, _ := s.Scan()
	if tok0 != tok {
		t.Errorf("%q: got %s, expected %s", src, tok0, tok)
	}
	if tok1 != token.EOF {
		t.Errorf("%q: got %s, expected EOF", src, tok1)
	}
	cnt := 0
	if err != "" {
		cnt = 1
	}
	if h.cnt != cnt {
		t.Errorf("%q: got cnt %d, expected %d", src, h.cnt, cnt)
	}
	if h.msg != err {
		t.Errorf("%q: got msg %q, expected %q", src, h.msg, err)
	}
	if h.pos.Offset != pos {
		t.Errorf("%q: got offset %d, expected %d", src, h.pos.Offset, pos)
	}
}


type srcerr struct {
	src string
	tok token.Token
	pos int
	err string
}

var errors = []srcerr{
	{"\"\"", token.STRING, 0, ""},
	{"\"", token.STRING, 0, "string not terminated"},
	{"/**/", token.COMMENT, 0, ""},
	{"/*", token.COMMENT, 0, "comment not terminated"},
	{"//\n", token.COMMENT, 0, ""},
	{"//", token.COMMENT, 0, "comment not terminated"},
	{"077", token.INT, 0, ""},
	{"078.", token.FLOAT, 0, ""},
	{"07801234567.", token.FLOAT, 0, ""},
	{"078e0", token.FLOAT, 0, ""},
	{"078", token.INT, 0, "illegal octal number"},
	{"07800000009", token.INT, 0, "illegal octal number"},
	{"\"abc\x00def\"", token.STRING, 4, "illegal character NUL"},
	{"\"abc\x80def\"", token.STRING, 4, "illegal UTF-8 encoding"},
}


func TestScanErrors(t *testing.T) {
	for _, e := range errors {
		checkError(t, e.src, e.tok, e.pos, e.err)
	}
}
