package cmd

import (
	"fmt"
	"strings"
)

type (
	AST struct {
		Properties Properties
		Sections   Sections
	}

	Sections map[string]Properties

	Properties map[string][]Value

	Value struct {
		String    string
		Variables []string
	}
)

func parseValues(tokens []Token) (Value, error) {
	var value Value
	var currentString strings.Builder

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		switch token.Type {
		case VAR:
			if tokens[i+1].Type == LCURLY && tokens[i+3].Type == RCURLY {
				varName := tokens[i+2]
				currentString.WriteString("%s")
				value.Variables = append(value.Variables, varName.Literal)
				i += 3
			} else {
				currentString.WriteString(token.Literal)
			}
		case ESCAPE:
			i++
			currentString.WriteString(tokens[i].Literal)
		default:
			currentString.WriteString(token.Literal)
		}
	}

	value.String = currentString.String()
	return value, nil
}

func GetAst(s string) (AST, error) {
	config := AST{Sections: make(Sections), Properties: make(Properties)}

	lexer := NewLexer(s)
	var currentSection string

	for tok := lexer.NextToken(); tok.Type != EOF; tok = lexer.NextToken() {

		if tok.Type != IDENT {
			continue
		}

		if lexer.lastToken.Type == LBRACKET {
			ending := lexer.NextToken()

			if ending.Type != RBRACKET {
				return config, fmt.Errorf("Invalid token at %+v %+v %+v", lexer.lastToken, tok, ending)
			}

			currentSection = tok.Literal
			config.Sections[tok.Literal] = make(Properties)
			continue
		}

		key := tok
		for lexer.NextToken().Type == WHITESPACE {
			// do nothing
		}
		assign := lexer.currentToken
		// Get the assign token
		for lexer.NextToken().Type == WHITESPACE {
			// do nothing
		}

		values := []Token{lexer.currentToken}

		for true {
			tok := lexer.NextToken()
			if tok.Type == EOF || tok.Type == EOL {
				break
			}
			values = append(values, tok)
		}

		if key.Type != IDENT || assign.Type != ASSIGN || len(values) == 0 {
			return config, fmt.Errorf("Invalid tokens %+v, %+v, %+v", key, assign, values)
		}

		value, err := parseValues(values)
		if err != nil {
			return config, err
		}

		if currentSection == "" {
			config.Properties[key.Literal] = append(config.Properties[key.Literal], value)
		} else {
			config.Sections[currentSection][key.Literal] = append(config.Sections[currentSection][key.Literal], value)
		}

	}

	return config, nil
}
