package cmd

import (
	"fmt"
	"strings"
)

type (
	AST struct {
		Properties AstProperties
		Sections   AstSections
	}

	AstSections   map[string]AstProperties
	AstProperties map[string][]AstValue

	AstValue struct {
		String    string
		Optional  bool
		Shell     map[int]string
		Variables map[int]string
	}
)

func parseSubShell(tokens []Token, index *int) (string, error) {
	varPose := *index
	varPose++ // Move to '('
	var shellString strings.Builder

	while := true
	for while {
		varPose++

		if varPose >= len(tokens) {
			return "", fmt.Errorf("Invalid token index for shell command: %d", varPose)
		}

		tok := tokens[varPose]
		switch tok.Type {
		case RPAREN:
			while = false
		case ESCAPE:
			varPose++
			shellString.WriteString(tokens[varPose].Literal)
		default:
			shellString.WriteString(tokens[varPose].Literal)
		}
	}

	*index = varPose

	return shellString.String(), nil
}

func parseValues(tokens []Token) (AstValue, error) {
	value := AstValue{
		Shell:     make(map[int]string),
		Variables: make(map[int]string),
	}

	var currentString strings.Builder

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		switch token.Type {
		case VAR:
			switch {
			case tokens[i+1].Type == LCURLY && tokens[i+3].Type == RCURLY:
				varName := tokens[i+2]
				value.Variables[currentString.Len()] = varName.Literal
				i += 3

			case tokens[i+1].Type == LPAREN:
				subShellString, err := parseSubShell(tokens, &i)
				if err != nil {
					return value, err
				}
				value.Shell[currentString.Len()] = subShellString

			default:
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
	config := AST{Sections: make(AstSections), Properties: make(AstProperties)}

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
			config.Sections[tok.Literal] = make(AstProperties)
			continue
		}

		key := tok
		for lexer.NextToken().Type == WHITESPACE {
			// do nothing
		}
		var assign Token
		optional := lexer.currentToken
		if optional.Type == OPTIONAL {
			assign = lexer.NextToken()
		} else {
			assign = optional
		}
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

		if currentSection != "" {
			config.Sections[currentSection][key.Literal] = append(config.Sections[currentSection][key.Literal], value)
			continue
		}

		if optional.Type != OPTIONAL {
			config.Properties[key.Literal] = append(config.Properties[key.Literal], value)
			continue
		}

		value.Optional = true
		config.Properties[key.Literal] = append(config.Properties[key.Literal], value)
	}

	return config, nil
}
