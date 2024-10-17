package cmd

import (
	"fmt"
)

type (
	AST struct {
		Properties Properties
		Sections   Sections
	}

	Sections map[string]Properties

	Properties map[string][]Values

	Values []Token
)

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

		values := Values{lexer.currentToken}

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

		if currentSection == "" {
			config.Properties[key.Literal] = append(config.Properties[key.Literal], values)
		} else {
			config.Sections[currentSection][key.Literal] = append(config.Sections[currentSection][key.Literal], values)
		}

	}

	return config, nil
}
