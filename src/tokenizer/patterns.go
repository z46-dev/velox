package tokenizer

import "regexp"

var cachedRegex map[TokenType]*regexp.Regexp

func GetRegexExpression(tokenType TokenType) string {
	switch tokenType {
	case Comment:
		return `^\/\/.*|^\/\*[\s\S]*?\*\/`
	case Preprocessor:
		return `^#\w+`
	case Number:
		return `^\d+(\.\d*)?`
	case String:
		return `^"[^"]*"`
	case Keyword:
		return `^(int|float|char|void|class|return|while|continue|break|if|else|New)\b`
	case Macro:
		return `^::`
	case Operator:
		return `^([+\-*/=<>!%&|^]=?|\+\+|--|&&|\|\||\*=|<<=?|>>=?|~|\?|:)`
	case Punctuation:
		return `^[{}()\[\];,.]`
	case Identifier:
		return `^[a-zA-Z_]\w*`
	case Whitespace:
		return `^\s+`
	default:
		return ""
	}
}

func GetRegex(tokenType TokenType) *regexp.Regexp {
	if cachedRegex == nil {
		cachedRegex = make(map[TokenType]*regexp.Regexp)
	}

	if cachedRegex[tokenType] == nil {
		cachedRegex[tokenType] = regexp.MustCompile(GetRegexExpression(tokenType))
	}

	return cachedRegex[tokenType]
}
