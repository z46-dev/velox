package tokenizer

type TokenType int

const (
	Invalid TokenType = iota
	Comment
	Preprocessor
	Number
	String
	Keyword
	Macro
	Operator
	Punctuation
	Identifier
	Whitespace
)

var TokenTypeNames map[TokenType]string = map[TokenType]string{
	Invalid:      "Invalid",
	Comment:      "Comment",
	Preprocessor: "Preprocessor",
	Number:       "Number",
	String:       "String",
	Keyword:      "Keyword",
	Macro:        "Macro",
	Operator:     "Operator",
	Punctuation:  "Punctuation",
	Identifier:   "Identifier",
	Whitespace:   "Whitespace",
}

type Token struct {
	Type         TokenType
	Value        string
	Line, Column int
}

func (token Token) String() string {
	return TokenTypeNames[token.Type] + "(" + token.Value + ")"
}
