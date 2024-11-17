package builder

import (
	"fmt"
	"strconv"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"velox.eparker.dev/src/ast"
)

type LoopTrace struct {
	condition, body, end *ir.Block
}

type Builder struct {
	ast             *ast.ASTNode
	module          *ir.Module
	functions       []*ir.Func
	currentFunction *ir.Func
	blocks          []*ir.Block
	currentBlock    *ir.Block
	locals          map[string]value.Value
	globals         map[string]constant.Constant
	loops           []*LoopTrace
}

func NewBuilder(ast *ast.ASTNode) *Builder {
	return &Builder{
		ast:       ast,
		module:    ir.NewModule(),
		locals:    make(map[string]value.Value),
		globals:   make(map[string]constant.Constant),
		loops:     make([]*LoopTrace, 0),
		functions: make([]*ir.Func, 0),
	}
}

type targetType int

const (
	Windows targetType = iota
	Linux
)

func (b *Builder) SetTarget(target targetType) *Builder {
	switch target {
	case Windows:
		b.module.TargetTriple = "x86_64-pc-windows-msvc"
	case Linux:
		b.module.TargetTriple = "x86_64-pc-linux-gnu"
	default:
		panic(fmt.Sprintf("Unsupported target: %d", target))
	}

	return b
}

func (b *Builder) Build() *ir.Module {
	for _, child := range b.ast.Children {
		switch child.Type {
		case ast.PreprocessorDirective:
			b.generatePreprocessorDirective(child)
		case ast.FunctionDeclaration:
			b.generateFunction(child)
		}
	}

	return b.module
}

func (b *Builder) generatePreprocessorDirective(node *ast.ASTNode) {
	if node.Name == "#define" && len(node.Children) == 2 {
		name := node.Children[0].Name

		// Parse value to int
		var value int64 = 0

		if node.Children[1].Type == ast.Literal {
			fmt.Sscanf(node.Children[1].Name, "%d", &value)
		} else {
			panic(fmt.Sprintf("Unsupported value type: %v", node.Children[1].Type))
		}

		// Add to globals
		b.globals[name] = constant.NewInt(types.I32, value)
		return
	}

	panic(fmt.Sprintf("Unsupported preprocessor directive: %v", node.Name))
}

func (b *Builder) generateFunction(node *ast.ASTNode) {
	name := node.Name
	retType := getTypeFromName(node.Children[0].Name)

	// Create parameter types
	var paramTypes []types.Type
	var paramNames []string
	for _, param := range node.Children[1].Children {
		paramName := param.Name
		paramType := getTypeFromName(param.Children[0].Name)
		paramTypes = append(paramTypes, paramType)
		paramNames = append(paramNames, paramName)
	}

	// Create function type
	var funcType *types.FuncType

	if len(paramTypes) > 0 {
		funcType = types.NewFunc(retType, paramTypes...)
	} else {
		funcType = types.NewFunc(retType)
	}

	var params []*ir.Param
	for i, paramName := range paramNames {
		params = append(params, ir.NewParam(paramName, paramTypes[i]))
	}

	// Create function
	fn := b.module.NewFunc(name, funcType.RetType, params...)

	// Add function to builder
	b.functions = append(b.functions, fn)
	b.currentFunction = fn

	// Locals
	b.locals = make(map[string]value.Value)
	for i, param := range fn.Params {
		param.SetName(paramNames[i])
		b.locals[paramNames[i]] = param
	}

	entry := fn.NewBlock("entry")
	b.blocks = append(b.blocks, entry)
	b.currentBlock = entry

	b.generateBlock(entry, node.Children[2])

	// Add return statement if not present
	if b.currentBlock.Term == nil {
		b.currentBlock.NewRet(nil)
	}

	b.currentFunction = nil
}

func (b *Builder) generateExpression(node *ast.ASTNode) value.Value {
	switch node.Type {
	case ast.Literal:
		return b.generateLiteral(node)
	case ast.Identifier:
		return b.generateIdentifier(node)
	case ast.BinaryExpression:
		return b.generateBinaryExpression(node)
	case ast.FunctionCall:
		return b.generateFunctionCall(node)
	default:
		panic(fmt.Sprintf("Unsupported expression type: %d", node.Type))
	}
}

func (b *Builder) generateLiteral(node *ast.ASTNode) value.Value {
	if val, err := strconv.Atoi(node.Name); err == nil {
		return constant.NewInt(types.I32, int64(val))
	} else if val, err := strconv.ParseFloat(node.Name, 64); err == nil {
		return constant.NewFloat(types.Double, val)
	}

	panic(fmt.Sprintf("Unsupported literal type: %s", node.Name))
}

func (b *Builder) generateIdentifier(node *ast.ASTNode) value.Value {
	if val, ok := b.locals[node.Name]; ok {
		if _, isParam := val.(*ir.Param); isParam {
			return val
		}

		return b.currentBlock.NewLoad(val.Type().(*types.PointerType).ElemType, val)
	}

	if val, ok := b.globals[node.Name]; ok {
		return val
	}

	panic(fmt.Sprintf("Unknown identifier: %s", node.Name))
}

func (b *Builder) generateBinaryExpression(node *ast.ASTNode) value.Value {
	left := b.generateExpression(node.Children[0])
	right := b.generateExpression(node.Children[1])

	lType, rType := left.Type(), right.Type()

	if lType != rType {
		panic(fmt.Sprintf("Binary expression types do not match: %v, %v", lType, rType))
	}

	if lType != types.I32 && lType != types.Double {
		panic(fmt.Sprintf("Unsupported binary expression type: %v", lType))
	}

	switch node.Name {
	case "+":
		switch lType {
		case types.I32:
			return b.currentBlock.NewAdd(left, right)
		case types.Double:
			return b.currentBlock.NewFAdd(left, right)
		}
	case "-":
		switch lType {
		case types.I32:
			return b.currentBlock.NewSub(left, right)
		case types.Double:
			return b.currentBlock.NewFSub(left, right)
		}
	case "*":
		switch lType {
		case types.I32:
			return b.currentBlock.NewMul(left, right)
		case types.Double:
			return b.currentBlock.NewFMul(left, right)
		}
	case "/":
		switch lType {
		case types.I32:
			return b.currentBlock.NewSDiv(left, right)
		case types.Double:
			return b.currentBlock.NewFDiv(left, right)
		}
	case "%":
		switch lType {
		case types.I32:
			return b.currentBlock.NewSRem(left, right)
		case types.Double:
			return b.currentBlock.NewFRem(left, right)
		}
	case "==":
		switch lType {
		case types.I32:
			return b.currentBlock.NewICmp(enum.IPredEQ, left, right)
		case types.Double:
			return b.currentBlock.NewFCmp(enum.FPredOEQ, left, right)
		}
	case "!=":
		switch lType {
		case types.I32:
			return b.currentBlock.NewICmp(enum.IPredNE, left, right)
		case types.Double:
			return b.currentBlock.NewFCmp(enum.FPredONE, left, right)
		}
	case "<":
		switch lType {
		case types.I32:
			return b.currentBlock.NewICmp(enum.IPredSLT, left, right)
		case types.Double:
			return b.currentBlock.NewFCmp(enum.FPredOLT, left, right)
		}
	case "<=":
		switch lType {
		case types.I32:
			return b.currentBlock.NewICmp(enum.IPredSLE, left, right)
		case types.Double:
			return b.currentBlock.NewFCmp(enum.FPredOLE, left, right)
		}
	case ">":
		switch lType {
		case types.I32:
			return b.currentBlock.NewICmp(enum.IPredSGT, left, right)
		case types.Double:
			return b.currentBlock.NewFCmp(enum.FPredOGT, left, right)
		}
	case ">=":
		switch lType {
		case types.I32:
			return b.currentBlock.NewICmp(enum.IPredSGE, left, right)
		case types.Double:
			return b.currentBlock.NewFCmp(enum.FPredOGE, left, right)
		}
	default:
		panic(fmt.Sprintf("Unsupported binary operator: %s", node.Name))
	}

	return nil
}

func (b *Builder) generateFunctionCall(node *ast.ASTNode) value.Value {
	fnName := node.Name
	var fn *ir.Func

	// Find the function in the module
	for _, f := range b.module.Funcs {
		if f.Name() == fnName {
			fn = f
			break
		}
	}

	if fn == nil {
		if fnName == "printf" {
			fn = b.module.NewFunc("printf", types.Void, ir.NewParam("format", types.NewPointer(types.I8)))
			fn.Sig.Variadic = true
		} else {
			panic(fmt.Sprintf("Function not found: %s", fnName))
		}
	}

	var args []value.Value
	for _, arg := range node.Children {
		args = append(args, b.generateExpression(arg))
	}

	if fnName == "printf" {
		formatStr := ""

		// Generate format string dynamically based on the argument types
		for _, arg := range args {
			switch arg.Type() {
			case types.I32:
				formatStr += "%d"
			case types.Double:
				formatStr += "%f"
			default:
				panic(fmt.Sprintf("Unsupported printf argument type: %v", arg.Type()))
			}
		}

		formatStr += "\n\x00"
		args = append([]value.Value{b.module.NewGlobalDef("", constant.NewCharArray([]byte(formatStr)))}, args...)
	}

	return b.currentBlock.NewCall(fn, args...)
}

func (b *Builder) generateBlock(block *ir.Block, node *ast.ASTNode) {
	b.currentBlock = block

	for _, child := range node.Children {
		switch child.Type {
		case ast.ReturnStatement:
			b.generateReturn(child)
		case ast.VariableDeclaration:
			b.generateVariableDeclaration(child)
		case ast.FunctionCall:
			b.generateFunctionCall(child)
		case ast.Statement:
			switch child.Name {
			case "if":
				b.generateConditional(child)
			case "continue":
				b.generateBreakContinue(true)
			case "break":
				b.generateBreakContinue(false)
			default:
				panic(fmt.Sprintf("Unsupported statement type: %s", child.Name))
			}
		case ast.Assignment:
			b.generateAssignment(child)
		case ast.WhileStatement:
			b.generateWhileStatement(child)
		default:
			panic(fmt.Sprintf("Unsupported block type: %s", ast.ASTNodeTypeNames[child.Type]))
		}
	}
}

func (b *Builder) generateReturn(node *ast.ASTNode) {
	if len(node.Children) == 0 {
		b.currentBlock.NewRet(nil)
		return
	}

	b.currentBlock.NewRet(b.generateExpression(node.Children[0]))
}

func (b *Builder) generateVariableDeclaration(node *ast.ASTNode) {
	name := node.Name

	alloca := b.currentBlock.NewAlloca(getTypeFromName(node.Children[0].Name))
	alloca.SetName(name)

	b.locals[name] = alloca

	if len(node.Children) > 1 {
		b.currentBlock.NewStore(b.generateExpression(node.Children[1]), alloca)
	}
}

func (b *Builder) generateAssignment(node *ast.ASTNode) {
	name := node.Name
	operator := node.Children[0].Name
	rightExpr := b.generateExpression(node.Children[1])

	alloca, ok := b.locals[name]

	if !ok {
		panic(fmt.Sprintf("Unknown identifier: %s", name))
	}

	if _, isParam := alloca.(*ir.Param); isParam {
		alloca = b.currentBlock.NewAlloca(alloca.Type())
		b.currentBlock.NewStore(b.locals[name], alloca)
	}

	loadInst := b.currentBlock.NewLoad(alloca.Type().(*types.PointerType).ElemType, alloca)

	var result value.Value

	switch operator {
	case "=":
		result = rightExpr
	case "+=":
		switch loadInst.Type() {
		case types.I32:
			result = b.currentBlock.NewAdd(loadInst, rightExpr)
		case types.Double:
			result = b.currentBlock.NewFAdd(loadInst, rightExpr)
		}
	case "-=":
		switch loadInst.Type() {
		case types.I32:
			result = b.currentBlock.NewSub(loadInst, rightExpr)
		case types.Double:
			result = b.currentBlock.NewFSub(loadInst, rightExpr)
		}
	case "*=":
		switch loadInst.Type() {
		case types.I32:
			result = b.currentBlock.NewMul(loadInst, rightExpr)
		case types.Double:
			result = b.currentBlock.NewFMul(loadInst, rightExpr)
		}
	case "/=":
		switch loadInst.Type() {
		case types.I32:
			result = b.currentBlock.NewSDiv(loadInst, rightExpr)
		case types.Double:
			result = b.currentBlock.NewFDiv(loadInst, rightExpr)
		}
	case "%=":
		switch loadInst.Type() {
		case types.I32:
			result = b.currentBlock.NewSRem(loadInst, rightExpr)
		case types.Double:
			result = b.currentBlock.NewFRem(loadInst, rightExpr)
		}
	default:
		panic(fmt.Sprintf("Unsupported assignment operator: %s", operator))
	}

	b.currentBlock.NewStore(result, alloca)
	b.locals[name] = alloca
}

func (b *Builder) generateConditional(node *ast.ASTNode) {
	// Generate the condition expression
	condition := b.generateExpression(node.Children[0])

	// Create basic blocks
	body := b.currentFunction.NewBlock(fmt.Sprintf("if.body.%d", len(b.blocks)))
	end := b.currentFunction.NewBlock(fmt.Sprintf("if.end.%d", len(b.blocks)))
	b.blocks = append(b.blocks, body, end)

	var elseBlock *ir.Block

	// Handle "else if" or "else"
	if len(node.Children) > 2 {
		if node.Children[2].Type == ast.Statement {
			// "else if" block
			elseBlock = b.currentFunction.NewBlock(fmt.Sprintf("if.elseif.%d", len(b.blocks)))
			b.blocks = append(b.blocks, elseBlock)
			b.currentBlock.NewCondBr(condition, body, elseBlock)
		} else {
			// "else" block
			elseBlock = b.currentFunction.NewBlock(fmt.Sprintf("if.else.%d", len(b.blocks)))
			b.blocks = append(b.blocks, elseBlock)
			b.currentBlock.NewCondBr(condition, body, elseBlock)
		}
	} else {
		// No "else" or "else if"
		b.currentBlock.NewCondBr(condition, body, end)
	}

	// Generate "if" body
	b.generateBlock(body, node.Children[1])
	if b.currentBlock.Term == nil {
		b.currentBlock.NewBr(end)
	}

	// Generate "else if" or "else" block
	if elseBlock != nil {
		b.currentBlock = elseBlock
		if node.Children[2].Type == ast.Statement {
			// Recursively handle "else if"
			b.generateConditional(node.Children[2])
		} else {
			// Handle "else"
			b.generateBlock(elseBlock, node.Children[2])
			if b.currentBlock.Term == nil {
				b.currentBlock.NewBr(end)
			}
		}
	}

	// Ensure "end" block has a terminator
	if b.currentBlock.Term == nil {
		b.currentBlock.NewBr(end)
	}
	b.currentBlock = end
}

func (b *Builder) generateWhileStatement(node *ast.ASTNode) {
	condition := b.currentFunction.NewBlock(fmt.Sprintf("while.cond.%d", len(b.blocks)))
	body := b.currentFunction.NewBlock(fmt.Sprintf("while.body.%d", len(b.blocks)))
	end := b.currentFunction.NewBlock(fmt.Sprintf("while.end.%d", len(b.blocks)))

	b.blocks = append(b.blocks, condition, body, end)

	b.currentBlock.NewBr(condition)

	// Add loop trace
	b.loops = append(b.loops, &LoopTrace{
		condition: condition,
		body:      body,
		end:       end,
	})

	// Generate condition block
	b.currentBlock = condition
	conditionExpr := b.generateExpression(node.Children[0])
	b.currentBlock.NewCondBr(conditionExpr, body, end)

	// Generate body block
	b.generateBlock(body, node.Children[1])
	if b.currentBlock.Term == nil {
		b.currentBlock.NewBr(condition)
	}

	// Ensure "end" block has a terminator
	if b.currentBlock.Term == nil {
		b.currentBlock.NewBr(end)
	}

	b.currentBlock = end
}

func (b *Builder) generateBreakContinue(isContinue bool) {
	if len(b.loops) == 0 {
		panic("Break/continue statement outside of loop")
	}

	var target *ir.Block

	if isContinue {
		target = b.loops[len(b.loops)-1].condition
	} else {
		target = b.loops[len(b.loops)-1].end
	}

	b.currentBlock.NewBr(target)
}

func getTypeFromName(name string) types.Type {
	switch name {
	case "int":
		return types.I32
	case "float":
		return types.Double
	case "void":
		return types.Void
	default:
		panic(fmt.Sprintf("Unsupported type: %s", name))
	}
}
