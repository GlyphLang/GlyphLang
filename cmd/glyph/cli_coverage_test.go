package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validSource is a minimal valid .glyph source with brace syntax
const validSource = `@ GET /hello {
  > {text: "Hello, World!"}
}
`

// --- Compile + Decompile round-trip ---

func TestCompileDecompileRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	srcFile := filepath.Join(tmpDir, "test.glyph")
	err := os.WriteFile(srcFile, []byte(validSource), 0644)
	require.NoError(t, err)

	// Compile
	compiledFile := filepath.Join(tmpDir, "test.glyphc")
	cmd := &cobra.Command{}
	cmd.Flags().String("output", compiledFile, "")
	cmd.Flags().Uint8("opt-level", 2, "")
	err = runCompile(cmd, []string{srcFile})
	require.NoError(t, err)
	assert.FileExists(t, compiledFile)

	// Verify compiled file is non-empty
	info, err := os.Stat(compiledFile)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))

	// Decompile
	decompFile := filepath.Join(tmpDir, "decompiled.glyph")
	cmd2 := &cobra.Command{}
	cmd2.Flags().String("output", decompFile, "")
	cmd2.Flags().Bool("disasm", false, "")
	err = runDecompile(cmd2, []string{compiledFile})
	require.NoError(t, err)
	assert.FileExists(t, decompFile)
}

func TestCompileDefaultOutput(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	srcFile := filepath.Join(tmpDir, "app.glyph")
	err := os.WriteFile(srcFile, []byte(validSource), 0644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("output", "", "")
	cmd.Flags().Uint8("opt-level", 0, "")
	err = runCompile(cmd, []string{srcFile})
	require.NoError(t, err)

	expectedOutput := filepath.Join(tmpDir, "app.glyphc")
	assert.FileExists(t, expectedOutput)
}

func TestCompileOptLevels(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.glyph")
	err := os.WriteFile(srcFile, []byte(validSource), 0644)
	require.NoError(t, err)

	for _, level := range []uint8{0, 1, 2, 3} {
		t.Run("opt-level-"+string(rune('0'+level)), func(t *testing.T) {
			outFile := filepath.Join(tmpDir, "out.glyphc")
			cmd := &cobra.Command{}
			cmd.Flags().String("output", outFile, "")
			cmd.Flags().Uint8("opt-level", level, "")
			err := runCompile(cmd, []string{srcFile})
			require.NoError(t, err)
			os.Remove(outFile)
		})
	}
}

// --- Compile error cases ---

func TestCompileNonExistentFile(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("output", "", "")
	cmd.Flags().Uint8("opt-level", 2, "")
	err := runCompile(cmd, []string{"/tmp/does-not-exist.glyph"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestCompileMalformedInput(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "bad.glyph")
	err := os.WriteFile(srcFile, []byte("@@@ this is not valid glyph"), 0644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("output", "", "")
	cmd.Flags().Uint8("opt-level", 2, "")
	err = runCompile(cmd, []string{srcFile})
	assert.Error(t, err)
}

func TestCompileNoRoutes(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "empty.glyph")
	// Valid parse but no routes
	err := os.WriteFile(srcFile, []byte(""), 0644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("output", "", "")
	cmd.Flags().Uint8("opt-level", 2, "")
	err = runCompile(cmd, []string{srcFile})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no routes found")
}

// --- Decompile error cases ---

func TestDecompileNonExistentFile(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("output", "", "")
	cmd.Flags().Bool("disasm", false, "")
	err := runDecompile(cmd, []string{"/tmp/does-not-exist.glyphc"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestDecompileInvalidBytecode(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad.glyphc")
	err := os.WriteFile(badFile, []byte("not bytecode"), 0644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("output", "", "")
	cmd.Flags().Bool("disasm", false, "")
	err = runDecompile(cmd, []string{badFile})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decompilation failed")
}

func TestDecompileDisasmOnly(t *testing.T) {
	tmpDir := t.TempDir()

	srcFile := filepath.Join(tmpDir, "test.glyph")
	err := os.WriteFile(srcFile, []byte(validSource), 0644)
	require.NoError(t, err)

	compiledFile := filepath.Join(tmpDir, "test.glyphc")
	cmd := &cobra.Command{}
	cmd.Flags().String("output", compiledFile, "")
	cmd.Flags().Uint8("opt-level", 0, "")
	err = runCompile(cmd, []string{srcFile})
	require.NoError(t, err)

	// Decompile with disasm-only flag
	cmd2 := &cobra.Command{}
	cmd2.Flags().String("output", "", "")
	cmd2.Flags().Bool("disasm", true, "")
	err = runDecompile(cmd2, []string{compiledFile})
	require.NoError(t, err)
}

// --- Init command ---

func TestRunInitHelloWorld(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	cmd := &cobra.Command{}
	cmd.Flags().String("template", "hello-world", "")
	err := runInit(cmd, []string{"my-project"})
	require.NoError(t, err)

	mainFile := filepath.Join(tmpDir, "my-project", "main.glyph")
	assert.FileExists(t, mainFile)

	content, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Hello World")
}

func TestRunInitRestAPI(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	cmd := &cobra.Command{}
	cmd.Flags().String("template", "rest-api", "")
	err := runInit(cmd, []string{"api-project"})
	require.NoError(t, err)

	mainFile := filepath.Join(tmpDir, "api-project", "main.glyph")
	assert.FileExists(t, mainFile)

	content, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "REST API")
}

func TestRunInitUnknownTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	cmd := &cobra.Command{}
	cmd.Flags().String("template", "nonexistent", "")
	err := runInit(cmd, []string{"bad-project"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown template")
}

// --- Validate command ---

func TestValidateFile_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "valid.glyph")
	err := os.WriteFile(srcFile, []byte(validSource), 0644)
	require.NoError(t, err)

	result := validateFile(srcFile)
	assert.True(t, result.Valid)
}

func TestValidateFile_NonExistent(t *testing.T) {
	result := validateFile("/tmp/does-not-exist-12345.glyph")
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
	assert.Equal(t, "file_error", result.Errors[0].Type)
}

func TestRunValidateCommand(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.glyph")
	err := os.WriteFile(srcFile, []byte(validSource), 0644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().Bool("ai", false, "")
	cmd.Flags().Bool("strict", false, "")
	cmd.Flags().Bool("quiet", false, "")
	err = runValidate(cmd, []string{srcFile})
	require.NoError(t, err)
}

func TestRunValidateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "a.glyph")
	err := os.WriteFile(srcFile, []byte(validSource), 0644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().Bool("ai", false, "")
	cmd.Flags().Bool("strict", false, "")
	cmd.Flags().Bool("quiet", false, "")
	err = runValidate(cmd, []string{tmpDir})
	require.NoError(t, err)
}

func TestRunValidateNonExistentPath(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("ai", false, "")
	cmd.Flags().Bool("strict", false, "")
	cmd.Flags().Bool("quiet", false, "")
	err := runValidate(cmd, []string{"/tmp/does-not-exist-xyz"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to access path")
}

func TestRunValidate_AIFlag(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.glyph")
	err := os.WriteFile(srcFile, []byte(validSource), 0644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().Bool("ai", true, "")
	cmd.Flags().Bool("strict", false, "")
	cmd.Flags().Bool("quiet", false, "")
	err = runValidate(cmd, []string{srcFile})
	require.NoError(t, err)
}

// --- Test command ---

func TestRunTestCommand_NoTests(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "no-tests.glyph")
	err := os.WriteFile(srcFile, []byte(validSource), 0644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().String("filter", "", "")
	cmd.Flags().Bool("fail-fast", false, "")
	err = runTest(cmd, []string{srcFile})
	require.NoError(t, err)
}

func TestRunTestCommand_NonExistentFile(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().String("filter", "", "")
	cmd.Flags().Bool("fail-fast", false, "")
	err := runTest(cmd, []string{"/tmp/does-not-exist.glyph"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

// --- Commands listing ---

func TestRunListCommands_NoCommands(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "routes-only.glyph")
	err := os.WriteFile(srcFile, []byte(validSource), 0644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	err = runListCommands(cmd, []string{srcFile})
	require.NoError(t, err)
}

func TestRunListCommands_NonExistentFile(t *testing.T) {
	cmd := &cobra.Command{}
	err := runListCommands(cmd, []string{"/tmp/does-not-exist.glyph"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

// --- Exec command error cases ---

func TestRunExec_NonExistentFile(t *testing.T) {
	cmd := &cobra.Command{}
	err := runExec(cmd, []string{"/tmp/does-not-exist.glyph", "hello"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestRunExec_CommandNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.glyph")
	err := os.WriteFile(srcFile, []byte(validSource), 0644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	err = runExec(cmd, []string{srcFile, "nonexistent"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no commands found")
}

// --- parseCommandArgs ---

func TestParseCommandArgs_FlagArgs(t *testing.T) {
	params := []ast.CommandParam{
		{Name: "name", IsFlag: true},
	}

	args := []string{"--name", "Alice"}
	result := parseCommandArgs(args, params)
	assert.Equal(t, "Alice", result["name"])
}

func TestParseCommandArgs_FlagEqualsFormat(t *testing.T) {
	params := []ast.CommandParam{
		{Name: "name", IsFlag: true},
	}

	args := []string{"--name=Bob"}
	result := parseCommandArgs(args, params)
	assert.Equal(t, "Bob", result["name"])
}

func TestParseCommandArgs_PositionalArgs(t *testing.T) {
	params := []ast.CommandParam{
		{Name: "input", IsFlag: false},
	}

	args := []string{"myfile.txt"}
	result := parseCommandArgs(args, params)
	assert.Equal(t, "myfile.txt", result["input"])
}

func TestParseCommandArgs_Empty(t *testing.T) {
	result := parseCommandArgs([]string{}, nil)
	assert.Empty(t, result)
}

func TestParseCommandArgs_MultipleFlags(t *testing.T) {
	params := []ast.CommandParam{
		{Name: "name", IsFlag: true},
		{Name: "age", IsFlag: true},
	}

	args := []string{"--name", "Alice", "--age", "30"}
	result := parseCommandArgs(args, params)
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, "30", result["age"])
}

// --- indexOf ---

func TestIndexOf(t *testing.T) {
	assert.Equal(t, 3, indexOf("foo=bar", '='))
	assert.Equal(t, -1, indexOf("foobar", '='))
	assert.Equal(t, 0, indexOf("=bar", '='))
	assert.Equal(t, -1, indexOf("", '='))
}

// --- convertHTTPMethod additional cases ---

func TestConvertHTTPMethod_Unknown(t *testing.T) {
	result := convertHTTPMethod(ast.HttpMethod(99))
	assert.Equal(t, "GET", string(result))
}

// --- parseSource cases ---

func TestParseSource_Empty(t *testing.T) {
	module, err := parseSource("")
	require.NoError(t, err)
	assert.NotNil(t, module)
	assert.Empty(t, module.Items)
}

func TestParseSource_MultipleRoutes(t *testing.T) {
	source := `@ GET /hello {
  > {text: "hello"}
}

@ POST /data {
  > {status: "created"}
}
`
	module, err := parseSource(source)
	require.NoError(t, err)
	assert.NotNil(t, module)
	assert.Len(t, module.Items, 2)
}

// --- Run command with bytecode ---

func TestRunBytecode(t *testing.T) {
	tmpDir := t.TempDir()

	srcFile := filepath.Join(tmpDir, "test.glyph")
	err := os.WriteFile(srcFile, []byte(validSource), 0644)
	require.NoError(t, err)

	compiledFile := filepath.Join(tmpDir, "test.glyphc")
	cmd := &cobra.Command{}
	cmd.Flags().String("output", compiledFile, "")
	cmd.Flags().Uint8("opt-level", 0, "")
	err = runCompile(cmd, []string{srcFile})
	require.NoError(t, err)

	// Run the bytecode
	runCmd := &cobra.Command{}
	runCmd.Flags().Uint16("port", 0, "")
	runCmd.Flags().Bool("bytecode", true, "")
	runCmd.Flags().Bool("interpret", false, "")
	err = runRun(runCmd, []string{compiledFile})
	require.NoError(t, err)
}

func TestRunBytecode_NonExistentFile(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Uint16("port", 0, "")
	cmd.Flags().Bool("bytecode", true, "")
	cmd.Flags().Bool("interpret", false, "")
	err := runRun(cmd, []string{"/tmp/does-not-exist.glyphc"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read bytecode file")
}

func TestRunBytecode_InvalidBytecode(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad.glyphc")
	err := os.WriteFile(badFile, []byte("not bytecode"), 0644)
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().Uint16("port", 0, "")
	cmd.Flags().Bool("bytecode", true, "")
	cmd.Flags().Bool("interpret", false, "")
	err = runRun(cmd, []string{badFile})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bytecode execution failed")
}

func TestRunAutoDetectBytecode(t *testing.T) {
	tmpDir := t.TempDir()

	srcFile := filepath.Join(tmpDir, "test.glyph")
	err := os.WriteFile(srcFile, []byte(validSource), 0644)
	require.NoError(t, err)

	compiledFile := filepath.Join(tmpDir, "test.glyphc")
	cmd := &cobra.Command{}
	cmd.Flags().String("output", compiledFile, "")
	cmd.Flags().Uint8("opt-level", 0, "")
	err = runCompile(cmd, []string{srcFile})
	require.NoError(t, err)

	// Run without --bytecode flag, should auto-detect from .glyphc extension
	runCmd := &cobra.Command{}
	runCmd.Flags().Uint16("port", 0, "")
	runCmd.Flags().Bool("bytecode", false, "")
	runCmd.Flags().Bool("interpret", false, "")
	err = runRun(runCmd, []string{compiledFile})
	require.NoError(t, err)
}
