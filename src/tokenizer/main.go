package tokenizer

import (
	"fmt"
	"os"
)

func lineCol(code string, position int) (int, int) {
	line := 1
	col := 1

	for i := 0; i < position; i++ {
		if code[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}

	return line, col
}

func Tokenize(code string, significantOnly bool) []Token {
	var tokens []Token
	var position int = 0

	for position < len(code) {
		var matched bool = false

		var line, col = lineCol(code, position)

		for tokenType := Comment; tokenType <= Whitespace; tokenType++ {
			expression := GetRegex(tokenType)

			if expression == nil {
				fmt.Printf("No expression for token type %d at %d:%d\n", tokenType, line, col)
				continue
			}

			match := expression.FindString(code[position:])

			if len(match) > 0 {
				if !significantOnly || (tokenType != Whitespace && tokenType != Comment) {
					tokens = append(tokens, Token{tokenType, match, line, col})
				}
				position += len(match)
				matched = true
				break
			}
		}

		if !matched {
			fmt.Printf("Unrecognized token at %d:%d\n", line, col)
			os.Exit(1)
		}
	}

	return tokens
}
