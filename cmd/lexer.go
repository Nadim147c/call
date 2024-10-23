package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type TokenType int

const (
	ILLEGAL    TokenType = iota
	EOF                  // End of file
	EOL                  // End of line
	IDENT                // Identifiers: variables, function names, etc.
	COMMENT              // "#" till the end of the line
	ASSIGN               // '='
	VAR                  // '$'
	ESCAPE               // '\'
	LPAREN               // '('
	RPAREN               // ')'
	LCURLY               // '{'
	RCURLY               // '}'
	LBRACKET             // '['
	RBRACKET             // ']'
	OPTIONAL             // '?'
	WHITESPACE           // '\s' '\t'
)

type Token struct {
	Type    TokenType
	Literal string
	Start   Position
	End     Position
}

type Position struct {
	Line   int
	Column int
}

type Lexer struct {
	scanner     *bufio.Scanner
	lineReader  *bufio.Reader
	currentLine string

	isLastRune bool
	lastRune   rune

	LastToken    *Token
	CurrentToken *Token

	Position     Position
	lastPosition Position
}

func NewLexer(s *bufio.Scanner) *Lexer {
	return &Lexer{scanner: s, Position: Position{Column: 0, Line: 0}}
}

func (l *Lexer) ReadRune() (rune, int, error) {
	l.lastPosition = l.Position

	line := l.Position.Line
	col := l.Position.Column

	if l.isLastRune {
		if l.lastRune == '\n' {
			col = 0
			line--
		}
		l.Position = Position{Column: col, Line: line}
		l.isLastRune = false
		return l.lastRune, utf8.RuneLen(l.lastRune), nil
	}

	if l.lineReader == nil {
		if !l.scanner.Scan() {
			if err := l.scanner.Err(); err != nil {
				return 0, 0, err
			}
			return 0, 0, io.EOF
		}
		col = 0
		line++
		l.currentLine = l.scanner.Text()
		if l.currentLine == "" {
			return l.ReadRune()
		}
		l.lineReader = bufio.NewReader(strings.NewReader(l.currentLine))
	}

	col++
	l.Position = Position{Column: col, Line: line}
	r, size, err := l.lineReader.ReadRune()
	if err == nil {
		l.lastRune = r
		return r, size, nil
	}

	l.lineReader = nil
	l.lastRune = '\n'
	return '\n', 1, nil
}

func (l *Lexer) UnreadRune() error {
	l.Position = l.lastPosition
	if l.isLastRune {
		return fmt.Errorf("no rune to un-read")
	}

	l.isLastRune = true
	return nil
}

func (l *Lexer) NextToken() Token {
	var token Token

	token.Start = l.Position

	l.LastToken = l.CurrentToken

	char, _, err := l.ReadRune()
	if err != nil {
		if err == io.EOF {
			return Token{Type: EOF}
		}
		return Token{Type: ILLEGAL}
	}

	switch char {
	case '=':
		token.Type = ASSIGN
	case '\\':
		token.Type = ESCAPE
	case '$':
		token.Type = VAR
	case '(':
		token.Type = LPAREN
	case ')':
		token.Type = RPAREN
	case '{':
		token.Type = LCURLY
	case '}':
		token.Type = RCURLY
	case '[':
		token.Type = LBRACKET
	case ']':
		token.Type = RBRACKET
	case '?':
		token.Type = OPTIONAL
	case ' ', '\t':
		token.Type = WHITESPACE
	case '\n':
		token.Type = EOL

	case '#':
		token.Type = COMMENT
		comment := string(char)
		for {
			r, _, err := l.ReadRune()
			if err != nil || r == '\n' {
				l.UnreadRune()
				break
			}
			comment += string(r)
		}
		token.Literal = string(comment)

	default:
		if unicode.IsLetter(char) {
			ident := string(char)
			for {
				r, _, err := l.ReadRune()
				if err != nil || !isIdent(r) {
					l.UnreadRune()
					break
				}
				ident += string(r)
			}

			token.Type = IDENT
			token.Literal = ident
			break
		}

		token.Type = ILLEGAL
	}

	if token.Literal == "" {
		token.Literal = string(char)
	}

	token.End = l.Position
	l.CurrentToken = &token

	return token
}

func isIdent(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-'
}
