package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"velox.eparker.dev/src/ast"
	"velox.eparker.dev/src/builder"
	"velox.eparker.dev/src/tokenizer"
)

func writeToJSONFile(filename string, data any) error {
	jsonData, err := json.MarshalIndent(data, "", "    ")

	if err != nil {
		return err
	}

	err = os.WriteFile(filename, jsonData, 0644)

	if err != nil {
		return err
	}

	return nil
}

func writeTextFile(filename string, data string) error {
	err := os.WriteFile(filename, []byte(data), 0644)

	if err != nil {
		return err
	}

	return nil
}

type Arguments struct {
	InputFile     string
	PreserveFiles bool
}

var args Arguments = (func() Arguments {
	var output Arguments

	for _, arg := range os.Args {
		switch arg {
		case "--preserve":
			output.PreserveFiles = true
		}

		if !strings.HasPrefix(arg, "--") {
			output.InputFile = arg
		}
	}

	if output.InputFile == "" {
		fmt.Println("No input file provided")
		os.Exit(1)
	}

	return output
})()

func checkOutputDir() {
	// Clear if exists, create if not
	if _, err := os.Stat("artifacts"); os.IsNotExist(err) {
		if err := os.Mkdir("artifacts", 0755); err != nil {
			fmt.Println("Error creating artifacts directory:", err)
		}
	} else {
		if err := os.RemoveAll("artifacts"); err != nil {
			fmt.Println("Error clearing artifacts directory:", err)
		}

		if err := os.Mkdir("artifacts", 0755); err != nil {
			fmt.Println("Error creating artifacts directory:", err)
		}
	}
}

func main() {
	file, err := os.OpenFile(args.InputFile, os.O_RDONLY, 0644)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	stat, err := file.Stat()

	if err != nil {
		panic(err)
	}

	buffer := make([]byte, stat.Size())

	_, err = file.Read(buffer)

	if err != nil {
		panic(err)
	}

	code := string(buffer)

	tokens := tokenizer.Tokenize(code, true)
	fmt.Printf("Found %d tokens\n", len(tokens))

	checkOutputDir()

	writeToJSONFile("./artifacts/tokens.json", tokens)

	ast := ast.NewParser(tokens, true).Parse()
	writeTextFile("./artifacts/ast.txt", ast.StringIndented(0))
	writeToJSONFile("./artifacts/ast.json", ast)

	writeTextFile("./artifacts/output.ll", builder.NewBuilder(ast).SetTarget(builder.Windows).Build().String())

	// Compile to assembly
	cmd := exec.Command("llc", "./artifacts/output.ll")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error compiling to assembly:")
		fmt.Println(string(output))
		panic(err)
	}

	// Compile to binary
	cmd = exec.Command("clang", "./artifacts/output.s")
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error compiling to binary:")
		fmt.Println(string(output))
		panic(err)
	}

	// Remove artifacts if not preserving
	if !args.PreserveFiles {
		if err := os.RemoveAll("artifacts"); err != nil {
			fmt.Println("Error removing artifacts directory:", err)
		}
	}

	fmt.Println("Done!")
}
