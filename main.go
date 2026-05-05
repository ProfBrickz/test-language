package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"lang-interpreter/interpreter"
	"lang-interpreter/lexer"
	"lang-interpreter/parser"
)

func main() {
	if len(os.Args) < 2 {
		repl()
		return
	}

	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
		os.Exit(1)
	}

	inter := interpreter.New()
	if err := inter.Run(program); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func repl() {
	inter := interpreter.New()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Welcome to the language interpreter (type 'exit' to quit)")
	fmt.Println()

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			fmt.Println()
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}

		l := lexer.New(line)
		p := parser.New(l)
		stmt, errs := p.ParseSingleStmt()
		if stmt == nil {
			if len(errs) > 0 {
				for _, err := range errs {
					fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				}
			}
			continue
		}
		if len(errs) > 0 {
			for _, err := range errs {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			}
			continue
		}

		if err := inter.ExecuteStmt(stmt); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}
	}
}
