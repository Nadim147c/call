package cmd

import (
	"fmt"
	"strings"
)

type (
	AST struct {
		Properties []AstGlobalProperty
		Sections   AstSections
	}

	AstGlobalProperty struct {
		Key   string
		Value AstValue
	}

	AstSections          map[string]AstSectionProperties
	AstSectionProperties map[string][]AstValue

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

		if token.Type == ESCAPE {
			i++
			currentString.WriteString(tokens[i].Literal)
			continue
		}

		if token.Type != VAR {
			currentString.WriteString(token.Literal)
			continue
		}

		if tokens[i+1].Type == LCURLY && tokens[i+3].Type == RCURLY {
			varName := tokens[i+2]
			value.Variables[currentString.Len()] = varName.Literal
			i += 3
		} else if tokens[i+1].Type == LPAREN {
			subShellString, err := parseSubShell(tokens, &i)
			if err != nil {
				return value, err
			}
			value.Shell[currentString.Len()] = subShellString
		} else {
			currentString.WriteString(token.Literal)
		}

	}

	value.String = currentString.String()
	return value, nil
}

func GetAst(s string) (AST, error) {
	config := AST{Sections: make(AstSections), Properties: []AstGlobalProperty{}}

	lexer := NewLexer(s)
	var currentSection string

	for tok := lexer.NextToken(); tok.Type != EOF; tok = lexer.NextToken() {
		if tok.Type == COMMENT {
			lexer.readComments()
		}

		if tok.Type != IDENT {
			continue
		}

		// Detect section
		if lexer.lastToken.Type == LBRACKET {
			ending := lexer.NextToken()

			if ending.Type != RBRACKET {
				return config, fmt.Errorf("Invalid token at %+v %+v %+v", lexer.lastToken, tok, ending)
			}

			currentSection = tok.Literal
			config.Sections[tok.Literal] = make(AstSectionProperties)

			for childTask := lexer.NextToken(); childTask.Type != EOL; childTask = lexer.NextToken() {
				if childTask.Type == IDENT {
					config.Sections[tok.Literal]["child"] = append(config.Sections[tok.Literal]["child"], AstValue{String: childTask.Literal})
				}

				if childTask.Type == COMMENT {
					lexer.readComments()
					break
				}
			}

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

		for lexer.NextToken().Type == WHITESPACE {
			// do nothing
		}

		values := []Token{lexer.currentToken}

		findValues := true
		for findValues {
			tok := lexer.NextToken()
			switch tok.Type {
			case EOF, EOL:
				findValues = false
			case COMMENT:
				lexer.readComments()
				findValues = false
			default:
				values = append(values, tok)
			}
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

		value.Optional = optional.Type == OPTIONAL
		config.Properties = append(config.Properties, AstGlobalProperty{Key: tok.Literal, Value: value})
	}

	return config, nil
}
