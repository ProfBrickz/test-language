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
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		repl()
		return nil
	}
	return runFile(os.Args[1])
}

func runFile(filename string) error {
	source, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Error reading file: %v", err)
	}

	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	for _, warn := range p.Warnings() {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", warn)
	}
	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
		return fmt.Errorf("parse errors")
	}

	inter := interpreter.New()
	if err := inter.Run(program); err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	return nil
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
		stmt, errs, warns := p.ParseSingleStmt()
		for _, warn := range warns {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", warn)
		}
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
