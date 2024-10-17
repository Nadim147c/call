package cmd

type TokenType int

const (
	ILLEGAL    TokenType = iota
	EOF                  // End of file
	EOL                  // End of line
	IDENT                // Identifiers: variables, function names, etc.
	INT                  // Integer literals
	ASSIGN               // '='
	VAR                  // '$'
	LPAREN               // '('
	RPAREN               // ')'
	LBRACKET             // '['
	RBRACKET             // ']'
	WHITESPACE           // \s \t
)

type Token struct {
	Type    TokenType
	Literal string
}

type Lexer struct {
	char         byte // current char under examination
	input        string
	lastToken    Token
	currentToken Token
	position     int // current position in input (points to current char)
	readPosition int // current reading position in input (after current char)
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.char = 0 // ASCII code for 'NUL' or end of input
	} else {
		l.char = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.lastToken = l.currentToken

	switch l.char {
	case '=':
		tok = Token{Type: ASSIGN, Literal: string(l.char)}
		l.currentToken = tok
	case '$':
		tok = Token{Type: VAR, Literal: string(l.char)}
		l.currentToken = tok
	case '(':
		tok = Token{Type: LPAREN, Literal: string(l.char)}
		l.currentToken = tok
	case ')':
		tok = Token{Type: RPAREN, Literal: string(l.char)}
		l.currentToken = tok
	case '[':
		tok = Token{Type: LBRACKET, Literal: string(l.char)}
		l.currentToken = tok
	case ']':
		tok = Token{Type: RBRACKET, Literal: string(l.char)}
		l.currentToken = tok
	case '\r':
		literal := "\r"
		if l.input[l.readPosition] == '\n' {
			l.readChar()
			literal += "\n"
		}
		tok = Token{Type: EOL, Literal: literal}
		tok = Token{Type: EOL, Literal: string(l.char)}
		l.currentToken = tok
	case '\n':
		tok = Token{Type: EOL, Literal: string(l.char)}
		l.currentToken = tok
	case ' ':
		tok = Token{Type: WHITESPACE, Literal: string(l.char)}
		l.currentToken = tok
	case '\t':
		tok = Token{Type: WHITESPACE, Literal: string(l.char)}
		l.currentToken = tok
	case 0:
		tok.Literal = ""
		tok.Type = EOF
		l.currentToken = tok
	default:
		switch {
		case isIdent(l.char):
			tok.Literal = l.readIdentifier()
			tok.Type = IDENT
			l.currentToken = tok
			return tok
		default:
			tok = Token{Type: ILLEGAL, Literal: string(l.char)}
			l.currentToken = tok
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isIdent(l.char) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(char byte) bool {
	return ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z')
}

func isDigit(char byte) bool {
	return '0' <= char && char <= '0'
}

func isIdent(char byte) bool {
	return isLetter(char) || isDigit(char) || (char == '_') || (char == '-')
}
