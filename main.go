package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"lang-interpreter/interpreter"
	"lang-interpreter/lexer"
	"lang-interpreter/parser"
	"lang-interpreter/static"
)

var osExit = os.Exit

const (
	colorRed    = "\x1b[31m"
	colorYellow = "\x1b[33m"
	colorReset  = "\x1b[0m"
)

type config struct {
	Errors   string `json:"errors"`
	Warnings string `json:"warnings"`
	Null     string `json:"null"`
}

func loadConfig() config {
	var cfg config
	data, err := os.ReadFile("lang-config.json")
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, &cfg)
	return cfg
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		osExit(1)
	}
}

func run() error {
	showErrors := interpreter.ShowBoth
	showWarnings := interpreter.ShowParse
	nullMode := static.NullWarn
	remainingArgs := []string{}

	cfg := loadConfig()
	if cfg.Errors != "" {
		switch cfg.Errors {
		case "parse":
			showErrors = interpreter.ShowParse
		case "run":
			showErrors = interpreter.ShowRun
		case "both":
			showErrors = interpreter.ShowBoth
		}
	}
	if cfg.Warnings != "" {
		switch cfg.Warnings {
		case "parse":
			showWarnings = interpreter.ShowParse
		case "run":
			showWarnings = interpreter.ShowRun
		case "both":
			showWarnings = interpreter.ShowBoth
		}
	}
	if cfg.Null != "" {
		switch cfg.Null {
		case "error":
			nullMode = static.NullError
		case "warning":
			nullMode = static.NullWarn
		case "none":
			nullMode = static.NullNone
		}
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--errors" || strings.HasPrefix(args[i], "--errors="):
			val := ""
			if idx := strings.Index(args[i], "="); idx >= 0 {
				val = args[i][idx+1:]
			} else if i+1 < len(args) {
				i++
				val = args[i]
			} else {
				return fmt.Errorf("--errors requires a value (parse, run, or both)")
			}
			switch val {
			case "parse":
				showErrors = interpreter.ShowParse
			case "run":
				showErrors = interpreter.ShowRun
			case "both":
				showErrors = interpreter.ShowBoth
			default:
				return fmt.Errorf("invalid value for --errors: %s (expected parse, run, or both)", val)
			}
		case args[i] == "--warnings" || strings.HasPrefix(args[i], "--warnings="):
			val := ""
			if idx := strings.Index(args[i], "="); idx >= 0 {
				val = args[i][idx+1:]
			} else if i+1 < len(args) {
				i++
				val = args[i]
			} else {
				return fmt.Errorf("--warnings requires a value (parse, run, or both)")
			}
			switch val {
			case "parse":
				showWarnings = interpreter.ShowParse
			case "run":
				showWarnings = interpreter.ShowRun
			case "both":
				showWarnings = interpreter.ShowBoth
			default:
				return fmt.Errorf("invalid value for --warnings: %s (expected parse, run, or both)", val)
			}
		case args[i] == "--null" || strings.HasPrefix(args[i], "--null="):
			val := ""
			if idx := strings.Index(args[i], "="); idx >= 0 {
				val = args[i][idx+1:]
			} else if i+1 < len(args) {
				i++
				val = args[i]
			} else {
				return fmt.Errorf("--null requires a value (error, warning, or none)")
			}
			switch val {
			case "error":
				nullMode = static.NullError
			case "warning":
				nullMode = static.NullWarn
			case "none":
				nullMode = static.NullNone
			default:
				return fmt.Errorf("invalid value for --null: %s (expected error, warning, or none)", val)
			}
		case args[i] == "--help" || args[i] == "-h":
			printHelp()
			return nil
		default:
			remainingArgs = append(remainingArgs, args[i])
		}
	}

	inter := interpreter.New()
	inter.SetShowErrors(showErrors)
	inter.SetShowWarnings(showWarnings)

	if len(remainingArgs) == 0 {
		replWith(inter, showErrors, showWarnings)
		return nil
	}
	return runFileWith(remainingArgs[0], inter, showErrors, showWarnings, nullMode)
}

func printHelp() {
	fmt.Println("test-language interpreter")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  test-language                          Start the REPL (interactive mode)")
	fmt.Println("  test-language <file>                   Execute a source file")
	fmt.Println("  test-language --help, -h               Show this help message")
	fmt.Println("  test-language --errors <mode>          Control when errors appear")
	fmt.Println("  test-language --warnings <mode>        Control when warnings appear")
	fmt.Println("  test-language --null <mode>            Control null-operation messages")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --errors [value]      Control when errors appear (default: both)")
	fmt.Println("      parse      Show at parse time only")
	fmt.Println("      run        Show at run time only")
	fmt.Println("      both       Show at both parse and run time")
	fmt.Println("  --warnings [value]    Control when warnings appear (default: parse)")
	fmt.Println("      parse      Show at parse time only")
	fmt.Println("      run        Show at run time only")
	fmt.Println("      both       Show at both parse and run time")
	fmt.Println("  --null [value]        Control null-operation detection (default: warning)")
	fmt.Println("      error      Report as error")
	fmt.Println("      warning    Report as warning")
	fmt.Println("      none       Suppress")
	fmt.Println("  --help, -h            Show this help message")
	fmt.Println()
	fmt.Println("Documentation: docs/ directory")
}

func runFile(filename string) error {
	return runFileWith(filename, interpreter.New(), interpreter.ShowBoth, interpreter.ShowBoth, static.NullWarn)
}

type lineMsg struct {
	line int
	text string
}

func parseLine(msg string) int {
	if idx := strings.Index(msg, "line "); idx >= 0 {
		rest := msg[idx+5:]
		if colon := strings.Index(rest, ":"); colon >= 0 {
			n, _ := strconv.Atoi(rest[:colon])
			return n
		}
	}
	return 0
}

func printLineOrder(msgs []lineMsg, source string) {
	if len(msgs) == 0 {
		return
	}

	var lines []string
	if source != "" {
		lines = strings.Split(source, "\n")
	}

	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].line < msgs[j].line
	})

	errCount := 0
	warnCount := 0

	for _, m := range msgs {
		fmt.Fprintf(os.Stderr, "%s\n", m.text)

		if strings.HasPrefix(m.text, colorRed) {
			errCount++
		} else if strings.HasPrefix(m.text, colorYellow) {
			warnCount++
		}

		if m.line > 0 && m.line <= len(lines) {
			fmt.Fprintf(os.Stderr, "  %d | %s\n", m.line, lines[m.line-1])
		}
	}

	fmt.Fprintf(os.Stderr, "\n%s", colorReset)
	fmt.Fprintf(os.Stderr, "Found %d error(s), %d warning(s)\n", errCount, warnCount)
}

func runFileWith(filename string, inter *interpreter.Interpreter, showErrors, showWarnings interpreter.ShowWhen, nullMode static.NullMode) error {
	source, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Error reading file: %v", err)
	}

	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	var msgs []lineMsg

	if inter.ShowParseWarnings() {
		for _, warn := range p.Warnings() {
			msgs = append(msgs, lineMsg{parseLine(warn), fmt.Sprintf("%sWarning:%s %s", colorYellow, colorReset, warn)})
		}
	}
	if inter.ShowParseErrors() {
		for _, err := range p.Errors() {
			msgs = append(msgs, lineMsg{parseLine(err), fmt.Sprintf("%sError:%s %s", colorRed, colorReset, err)})
		}
	}
	if len(p.Errors()) > 0 {
		printLineOrder(msgs, string(source))
		return fmt.Errorf("parse errors")
	}

	useStaticErrors := showErrors == interpreter.ShowParse || showErrors == interpreter.ShowBoth
	useStaticWarnings := showWarnings == interpreter.ShowParse || showWarnings == interpreter.ShowBoth

	if useStaticErrors || useStaticWarnings {
		analyzer := static.New()
		analyzer.SetNullMode(nullMode)
		analyzer.Analyze(program)

		if useStaticWarnings {
			for _, warn := range analyzer.Warnings() {
				msgs = append(msgs, lineMsg{parseLine(warn), fmt.Sprintf("%sWarning:%s %s", colorYellow, colorReset, warn)})
			}
		}
		if useStaticErrors {
			for _, err := range analyzer.Errors() {
				msgs = append(msgs, lineMsg{parseLine(err), fmt.Sprintf("%sError:%s %s", colorRed, colorReset, err)})
			}
			if len(analyzer.Errors()) > 0 {
				printLineOrder(msgs, string(source))
				return fmt.Errorf("static analysis found errors")
			}
		}
		printLineOrder(msgs, string(source))
	}

	useRuntimeErrors := showErrors == interpreter.ShowRun || showErrors == interpreter.ShowBoth
	if useRuntimeErrors {
		if err := inter.Run(program); err != nil {
			msg := fmt.Sprintf("%sError:%s %s", colorRed, colorReset, err.Error())
			msgs = append(msgs, lineMsg{parseLine(err.Error()), msg})
			printLineOrder(msgs, string(source))
			return fmt.Errorf("runtime error")
		}
	}
	return nil
}

func repl() {
	replWith(interpreter.New(), interpreter.ShowBoth, interpreter.ShowBoth)
}

func replWith(inter *interpreter.Interpreter, showErrors, showWarnings interpreter.ShowWhen) {
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
		var msgs []lineMsg
		if inter.ShowParseWarnings() {
			for _, warn := range warns {
				msgs = append(msgs, lineMsg{parseLine(warn), fmt.Sprintf("%sWarning:%s %s", colorYellow, colorReset, warn)})
			}
		}
		if inter.ShowParseErrors() {
			for _, err := range errs {
				msgs = append(msgs, lineMsg{parseLine(err), fmt.Sprintf("%sError:%s %s", colorRed, colorReset, err)})
			}
		}
		printLineOrder(msgs, "")
		if stmt == nil {
			continue
		}
		if len(errs) > 0 {
			continue
		}

		useRuntimeErrors := showErrors == interpreter.ShowRun || showErrors == interpreter.ShowBoth
		if useRuntimeErrors {
			if err := inter.ExecuteStmt(stmt); err != nil {
				fmt.Fprintf(os.Stderr, "%sError:%s %v\n", colorRed, colorReset, err)
				continue
			}
		}
	}
}
