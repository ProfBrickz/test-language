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
		inW.WriteString("var x: integer{size: 32} = 42;\nexit\n")
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
	content := "var x: integer{size: 32} = 42;"
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
	content := "var x: integer{size: 32} = 42/0;"
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
