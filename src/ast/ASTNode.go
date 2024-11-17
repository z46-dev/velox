package ast

import "fmt"

type ASTNode struct {
	Type     ASTNodeType
	Children []*ASTNode
	Name     string
}

func (node *ASTNode) String() string { // Tree representation
	var result string = fmt.Sprintf("%s(%s)", ASTNodeTypeNames[node.Type], node.Name)

	if len(node.Children) > 0 {
		result += " {"

		for _, child := range node.Children {
			result += " " + child.String()
		}

		result += " }"
	}

	return result
}

func (node *ASTNode) StringIndented(indentLevel int) string {
	var result string = fmt.Sprintf("%s(%s)", ASTNodeTypeNames[node.Type], node.Name)

	if len(node.Children) > 0 {
		result += " {\n"

		for _, child := range node.Children {
			for i := 0; i < indentLevel+1; i++ {
				result += "    "
			}

			result += child.StringIndented(indentLevel+1) + "\n"
		}

		for i := 0; i < indentLevel; i++ {
			result += "    "
		}

		result += "}"
	}

	return result
}
