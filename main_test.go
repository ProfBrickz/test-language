package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplWelcome(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w

	in, inW, _ := os.Pipe()
	os.Stdin = in
	go func() {
		inW.WriteString("exit\n")
		inW.Close()
	}()

	repl()

	w.Close()
	out, _ := io.ReadAll(r)

	if !strings.Contains(string(out), "Welcome to the language interpreter") {
		t.Errorf("expected welcome message, got: %s", string(out))
	}
}

func TestReplEmptyInput(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w

	in, inW, _ := os.Pipe()
	os.Stdin = in
	go func() {
		inW.WriteString("\n\n\nexit\n")
		inW.Close()
	}()

	repl()

	w.Close()
	out, _ := io.ReadAll(r)

	if strings.Contains(string(out), "Error:") {
		t.Errorf("should not have errors for empty input, got: %s", string(out))
	}
}

func TestReplQuit(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w

	in, inW, _ := os.Pipe()
	os.Stdin = in
	go func() {
		inW.WriteString("quit\n")
		inW.Close()
	}()

	repl()

	w.Close()
	out, _ := io.ReadAll(r)

	if strings.Contains(string(out), "Error:") {
		t.Errorf("should not have errors for quit, got: %s", string(out))
	}
}

func TestReplValidStatement(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w
	e, ew, _ := os.Pipe()
	os.Stderr = ew

	in, inW, _ := os.Pipe()
	os.Stdin = in
	go func() {
		inW.WriteString("var x: int{size: 32} = 42;\nexit\n")
		inW.Close()
	}()

	repl()

	w.Close()
	ew.Close()
	out, _ := io.ReadAll(r)
	errOut, _ := io.ReadAll(e)

	if strings.Contains(string(out)+string(errOut), "Error:") {
		t.Errorf("should not have errors, got: stdout=%s, stderr=%s", string(out), string(errOut))
	}
}

func TestReplParseError(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w
	e, ew, _ := os.Pipe()
	os.Stderr = ew

	in, inW, _ := os.Pipe()
	os.Stdin = in
	go func() {
		inW.WriteString("invalid syntax;\nexit\n")
		inW.Close()
	}()

	repl()

	w.Close()
	ew.Close()
	out, _ := io.ReadAll(r)
	errOut, _ := io.ReadAll(e)

	combined := string(out) + string(errOut)
	if !strings.Contains(combined, "Error:") {
		t.Errorf("expected error for invalid syntax, got: stdout=%s, stderr=%s", string(out), string(errOut))
	}
}

func TestReplSemicolonOnly(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w
	e, ew, _ := os.Pipe()
	os.Stderr = ew

	in, inW, _ := os.Pipe()
	os.Stdin = in
	go func() {
		inW.WriteString(";\nexit\n")
		inW.Close()
	}()

	repl()

	w.Close()
	ew.Close()
	out, _ := io.ReadAll(r)
	errOut, _ := io.ReadAll(e)

	combined := string(out) + string(errOut)
	if !strings.Contains(combined, "Error") {
		t.Errorf("expected error for semicolon-only input")
	}
}

func TestReplStmtWithErrors(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w
	e, ew, _ := os.Pipe()
	os.Stderr = ew

	in, inW, _ := os.Pipe()
	os.Stdin = in
	go func() {
		inW.WriteString("var x: int{size: 7} = 42;\nexit\n")
		inW.Close()
	}()

	repl()

	w.Close()
	ew.Close()
	out, _ := io.ReadAll(r)
	errOut, _ := io.ReadAll(e)

	combined := string(out) + string(errOut)
	if !strings.Contains(combined, "invalid size") {
		t.Errorf("expected 'invalid size' error, got: stdout=%s, stderr=%s", string(out), string(errOut))
	}
}

func TestReplExecutionError(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w
	e, ew, _ := os.Pipe()
	os.Stderr = ew

	in, inW, _ := os.Pipe()
	os.Stdin = in
	go func() {
		inW.WriteString("print(x);\nexit\n")
		inW.Close()
	}()

	repl()

	w.Close()
	ew.Close()
	out, _ := io.ReadAll(r)
	errOut, _ := io.ReadAll(e)

	combined := string(out) + string(errOut)
	if !strings.Contains(combined, "Error:") {
		t.Errorf("expected error for undefined variable, got: stdout=%s, stderr=%s", string(out), string(errOut))
	}
}

func TestRunMainWithFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 32} = 42;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"lang", testFile}
	err := run()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestRunFileParseErrors(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "invalid syntax here;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := runFile(testFile)
	if err == nil {
		t.Errorf("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse errors") {
		t.Errorf("expected 'parse errors' in error, got: %v", err)
	}
}

func TestRunFileRuntimeError(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 32} = 42/0;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := runFile(testFile)
	if err == nil {
		t.Errorf("expected runtime error")
	}
}

func TestRunMainWithInvalidFile(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"lang", "nonexistent.file"}
	err := run()
	if err == nil {
		t.Errorf("expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "Error reading file") {
		t.Errorf("expected 'Error reading file' in error, got: %v", err)
	}
}

func TestRunMainNoArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	in, inW, _ := os.Pipe()
	oldStdin := os.Stdin
	os.Stdin = in
	defer func() { os.Stdin = oldStdin }()

	go func() {
		inW.WriteString("exit\n")
		inW.Close()
	}()

	os.Args = []string{"lang"}
	err := run()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMainBinarySmokeTest(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "test-lang")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = "/home/Ryan/GitHub/test-language"
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}

	cmd = exec.Command(binPath)
	cmd.Stdin = strings.NewReader("exit\n")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !strings.Contains(out.String(), "Welcome to the language interpreter") {
		t.Errorf("expected welcome message, got: %s", out.String())
	}
}

func TestMainErrorExit(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "test-lang")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = "/home/Ryan/GitHub/test-language"
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}

	cmd = exec.Command(binPath, "nonexistent.file")
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()

	if err == nil {
		t.Errorf("expected exit error for non-existent file")
	}
	if !strings.Contains(errOut.String(), "Error reading file") {
		t.Errorf("expected 'Error reading file', got: %s", errOut.String())
	}
}

func TestRunFileWithWarning(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 32} = 010;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := runFile(testFile)

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(string(errOut), "Warning:") {
		t.Errorf("expected warning output, got: %s", string(errOut))
	}
}

func TestReplWithWarning(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w
	e, ew, _ := os.Pipe()
	os.Stderr = ew

	in, inW, _ := os.Pipe()
	os.Stdin = in
	go func() {
		inW.WriteString("print(010);\nexit\n")
		inW.Close()
	}()

	repl()

	w.Close()
	ew.Close()
	out, _ := io.ReadAll(r)
	errOut, _ := io.ReadAll(e)

	combined := string(out) + string(errOut)
	if !strings.Contains(combined, "Warning:") {
		t.Errorf("expected warning output, got: stdout=%s, stderr=%s", string(out), string(errOut))
	}
}

func TestMainErrorPath(t *testing.T) {
	exited := false
	exitCode := 0
	osExit = func(code int) {
		exited = true
		exitCode = code
		panic("os.Exit called")
	}
	defer func() { osExit = os.Exit }()

	oldArgs := os.Args
	os.Args = []string{"lang", "nonexistent.file"}
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	func() {
		defer func() {
			if r := recover(); r != nil {
				// expected panic from osExit
			}
		}()
		main()
	}()

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if !exited {
		t.Errorf("expected os.Exit to be called")
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(string(errOut), "Error reading file") {
		t.Errorf("expected 'Error reading file', got: %s", string(errOut))
	}
}

func TestHelpFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", "--help"}
	err := run()
	if err != nil {
		t.Errorf("expected no error for --help, got %v", err)
	}
}

func TestShortHelpFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", "-h"}
	err := run()
	if err != nil {
		t.Errorf("expected no error for -h, got %v", err)
	}
}

func TestWarningsParseFlag(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 32} = 010;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	os.Args = []string{"test-language", "--warnings", "parse", testFile}
	err := run()

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(string(errOut), "Warning:") {
		t.Errorf("expected warning with --warnings parse, got: %s", string(errOut))
	}
}

func TestWarningsRunFlagSuppressesWarnings(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 32} = 010;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	os.Args = []string{"test-language", "--warnings", "run", testFile}
	err := run()

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if strings.Contains(string(errOut), "Warning:") {
		t.Errorf("expected no warnings with --warnings run, got: %s", string(errOut))
	}
}

func TestWarningsBothFlag(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 32} = 010;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	os.Args = []string{"test-language", "--warnings", "both", testFile}
	err := run()

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(string(errOut), "Warning:") {
		t.Errorf("expected warning with --warnings both, got: %s", string(errOut))
	}
}

func TestErrorsRunFlag(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 32} = 42/0;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", "--errors", "run", testFile}
	err := run()
	if err == nil {
		t.Errorf("expected error for runtime error with --errors run")
	}
}

func TestErrorsParseFlag(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "invalid syntax here;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", "--errors", "parse", testFile}
	err := run()
	if err == nil {
		t.Errorf("expected parse error with --errors parse")
	}
}

func TestErrorsBothFlag(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 32} = 42;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", "--errors", "both", testFile}
	err := run()
	if err != nil {
		t.Errorf("expected no error with --errors both, got %v", err)
	}
}

func TestErrorsInvalidValue(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", "--errors", "invalid"}
	err := run()
	if err == nil {
		t.Errorf("expected error for invalid --errors value")
	}
}

func TestErrorsMissingValue(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", "--errors"}
	err := run()
	if err == nil {
		t.Errorf("expected error for --errors without value")
	}
}

func TestWarningsInvalidValue(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", "--warnings", "invalid"}
	err := run()
	if err == nil {
		t.Errorf("expected error for invalid --warnings value")
	}
}

func TestWarningsMissingValue(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", "--warnings"}
	err := run()
	if err == nil {
		t.Errorf("expected error for --warnings without value")
	}
}

func TestReplScannerEOF(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w

	in, inW, _ := os.Pipe()
	os.Stdin = in
	inW.Close()

	repl()

	w.Close()
	out, _ := io.ReadAll(r)

	if !strings.Contains(string(out), "Welcome to the language interpreter") {
		t.Errorf("expected welcome message, got: %s", string(out))
	}
}

func TestErrorsParseCatchesStaticError(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var a: int = 0; var b: int = 10 / a;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	os.Args = []string{"test-language", "--errors", "parse", testFile}
	err := run()

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if err == nil {
		t.Errorf("expected error for static analysis finding")
	}
	if !strings.Contains(string(errOut), "division by zero will occur") {
		t.Errorf("expected 'division by zero will occur', got: %s", string(errOut))
	}
}

func TestWarningsParseShowsStaticWarning(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 64, signed: true, nullable: true} = 10; var y: int = 10 / x;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	os.Args = []string{"test-language", "--errors", "run", "--warnings", "parse", testFile}
	err := run()

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(string(errOut), "value may be null when used with operator") {
		t.Errorf("expected null warning, got: %s", string(errOut))
	}
}

func TestParseLineNoLine(t *testing.T) {
	line := parseLine("no line number here")
	if line != 0 {
		t.Errorf("expected 0, got %d", line)
	}
}

func TestParseLineNoColon(t *testing.T) {
	line := parseLine("line 5 no colon after number")
	if line != 0 {
		t.Errorf("expected 0, got %d", line)
	}
}

func TestParseLineNonNumeric(t *testing.T) {
	line := parseLine("line abc: message")
	if line != 0 {
		t.Errorf("expected 0, got %d", line)
	}
}

func TestParseLineValid(t *testing.T) {
	line := parseLine("line 42: some message")
	if line != 42 {
		t.Errorf("expected 42, got %d", line)
	}
}

func TestNullFlagNone(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 64, signed: true, nullable: true} = 10; var y: int = 10 / x;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	os.Args = []string{"test-language", "--null", "none", testFile}
	err := run()

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if err != nil {
		t.Errorf("expected no error with --null none, got %v", err)
	}
	if strings.Contains(string(errOut), "Warning:") {
		t.Errorf("expected no warnings with --null none, got: %s", string(errOut))
	}
}

func TestNullFlagError(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 64, signed: true, nullable: true} = 10; var y: int = 10 / x;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", "--null", "error", testFile}
	err := run()
	if err == nil {
		t.Errorf("expected error with --null error")
	}
}

func TestNullFlagMissingValue(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", "--null"}
	err := run()
	if err == nil {
		t.Errorf("expected error for --null without value")
	}
}

func TestNullFlagInvalidValue(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", "--null", "invalid"}
	err := run()
	if err == nil {
		t.Errorf("expected error for invalid --null value")
	}
}

func TestNullFlagWarningValue(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 64, signed: true, nullable: true} = 10; var y: int = 10 / x;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	os.Args = []string{"test-language", "--null", "warning", testFile}
	err := run()

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if err != nil {
		t.Errorf("expected no error with --null warning, got %v", err)
	}
	if !strings.Contains(string(errOut), "Warning:") {
		t.Errorf("expected warning with --null warning, got: %s", string(errOut))
	}
}

func TestNullFlagEqualsSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 64, signed: true, nullable: true} = 10; var y: int = 10 / x;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	os.Args = []string{"test-language", "--null=none", testFile}
	err := run()

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if err != nil {
		t.Errorf("expected no error with --null=none, got %v", err)
	}
	if strings.Contains(string(errOut), "Warning:") {
		t.Errorf("expected no warnings with --null=none, got: %s", string(errOut))
	}
}

func TestErrorsFlagEqualsSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 64, signed: true, nullable: true} = 10; var y: int = 10 / x;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", "--errors=run", "--null=none", testFile}
	err := run()
	if err != nil {
		t.Errorf("expected no error with --errors=run --null=none, got %v", err)
	}
}

func TestWarningsFlagEqualsSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 64, signed: true, nullable: true} = 10; var y: int = 10 / x;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	os.Args = []string{"test-language", "--warnings=parse", "--null=none", testFile}
	err := run()

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if err != nil {
		t.Errorf("expected no error with --warnings=parse --null=none, got %v", err)
	}
	if strings.Contains(string(errOut), "Warning:") {
		t.Errorf("expected no warnings with --null=none, got: %s", string(errOut))
	}
}

func writeConfigFile(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "lang-config.json"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
}

func TestConfigFileApplied(t *testing.T) {
	tmpDir := t.TempDir()
	writeConfigFile(t, tmpDir, `{"null": "none", "warnings": "run"}`)

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 64, signed: true, nullable: true} = 10; var y: int = 10 / x;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	os.Args = []string{"test-language", testFile}
	err := run()

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if err != nil {
		t.Errorf("expected no error with config, got %v", err)
	}
	if strings.Contains(string(errOut), "Warning:") {
		t.Errorf("expected no warnings from config, got: %s", string(errOut))
	}
}

func TestCliOverridesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	writeConfigFile(t, tmpDir, `{"null": "none", "warnings": "run"}`)

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 64, signed: true, nullable: true} = 10; var y: int = 10 / x;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	os.Args = []string{"test-language", "--null=warning", "--warnings=both", testFile}
	err := run()

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !strings.Contains(string(errOut), "Warning:") {
		t.Errorf("expected warning from CLI override, got: %s", string(errOut))
	}
}

func TestConfigFileInvalidJson(t *testing.T) {
	tmpDir := t.TempDir()
	writeConfigFile(t, tmpDir, "not valid json")

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language"}
	err := run()
	if err != nil {
		t.Errorf("expected no error for invalid config, got %v", err)
	}
}

func TestConfigWarningsParse(t *testing.T) {
	tmpDir := t.TempDir()
	writeConfigFile(t, tmpDir, `{"warnings": "parse", "null": "warning"}`)

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 64, signed: true, nullable: true} = 10; var y: int = 10 / x;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	os.Args = []string{"test-language", testFile}
	err := run()

	w.Close()
	os.Stderr = oldStderr
	errOut, _ := io.ReadAll(r)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !strings.Contains(string(errOut), "Warning:") {
		t.Errorf("expected warning from config, got: %s", string(errOut))
	}
}

func TestConfigErrorsParse(t *testing.T) {
	tmpDir := t.TempDir()
	writeConfigFile(t, tmpDir, `{"errors": "parse", "null": "error"}`)

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 64, signed: true, nullable: true} = 10; var y: int = 10 / x;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", testFile}
	err := run()
	if err == nil {
		t.Errorf("expected error from config with null=error")
	}
}

func TestConfigErrorsBoth(t *testing.T) {
	tmpDir := t.TempDir()
	writeConfigFile(t, tmpDir, `{"errors": "both", "warnings": "both", "null": "none"}`)

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 64} = 10; var y: int = x + 1;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", testFile}
	err := run()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestConfigErrorsRun(t *testing.T) {
	tmpDir := t.TempDir()
	writeConfigFile(t, tmpDir, `{"errors": "run"}`)

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	testFile := filepath.Join(tmpDir, "test.lang")
	content := "var x: int{size: 64} = 10; var y: int = x + 1;"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test-language", testFile}
	err := run()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
