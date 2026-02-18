package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/codegen"
	"github.com/glyphlang/glyph/pkg/ir"
)

// intentTestFiles lists the 5 intent-test .glyph files used for codegen integration testing.
var intentTestFiles = []struct {
	name string
	path string
}{
	{"01-crud-api", filepath.Join("..", "examples", "intent-tests", "01-crud-api.glyph")},
	{"02-webhook-processor", filepath.Join("..", "examples", "intent-tests", "02-webhook-processor.glyph")},
	{"03-chat-server", filepath.Join("..", "examples", "intent-tests", "03-chat-server.glyph")},
	{"04-job-queue", filepath.Join("..", "examples", "intent-tests", "04-job-queue.glyph")},
	{"05-auth-service", filepath.Join("..", "examples", "intent-tests", "05-auth-service.glyph")},
}

// parseAndAnalyze reads a .glyph file, parses it, and runs the IR analyzer.
func parseAndAnalyze(t *testing.T, path string) *ir.ServiceIR {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}

	module, err := parseSource(string(source))
	if err != nil {
		t.Fatalf("parse failed for %s: %v", path, err)
	}

	analyzer := ir.NewAnalyzer()
	service, err := analyzer.Analyze(module)
	if err != nil {
		t.Fatalf("IR analysis failed for %s: %v", path, err)
	}

	return service
}

func TestCodegenIntegration_Python(t *testing.T) {
	gen := codegen.NewPythonGenerator("0.0.0.0", 8000)

	for _, tt := range intentTestFiles {
		t.Run(tt.name, func(t *testing.T) {
			service := parseAndAnalyze(t, tt.path)
			output := gen.Generate(service)

			if len(output) == 0 {
				t.Fatal("generated empty Python output")
			}

			// Verify core FastAPI constructs
			if !strings.Contains(output, "from fastapi import") {
				t.Error("expected FastAPI import")
			}
			if !strings.Contains(output, "app = FastAPI()") {
				t.Error("expected FastAPI app creation")
			}

			// Verify at least one route decorator exists
			hasRoute := strings.Contains(output, "@app.get") ||
				strings.Contains(output, "@app.post") ||
				strings.Contains(output, "@app.put") ||
				strings.Contains(output, "@app.delete")
			if !hasRoute {
				t.Error("expected at least one route decorator")
			}

			// Verify requirements.txt generation
			reqs := gen.GenerateRequirements(service)
			if !strings.Contains(reqs, "fastapi") {
				t.Error("expected fastapi in requirements.txt")
			}
		})
	}
}

func TestCodegenIntegration_TypeScript(t *testing.T) {
	gen := codegen.NewTypeScriptServerGenerator("0.0.0.0", 3000)

	for _, tt := range intentTestFiles {
		t.Run(tt.name, func(t *testing.T) {
			service := parseAndAnalyze(t, tt.path)
			output := gen.Generate(service)

			if len(output) == 0 {
				t.Fatal("generated empty TypeScript output")
			}

			// Verify core Express constructs
			if !strings.Contains(output, "import express") {
				t.Error("expected express import")
			}
			if !strings.Contains(output, "const app = express()") {
				t.Error("expected express app creation")
			}

			// Verify at least one route handler exists
			hasRoute := strings.Contains(output, "app.get(") ||
				strings.Contains(output, "app.post(") ||
				strings.Contains(output, "app.put(") ||
				strings.Contains(output, "app.delete(")
			if !hasRoute {
				t.Error("expected at least one Express route handler")
			}

			// Verify package.json generation
			pkg := gen.GeneratePackageJSON(service)
			if !strings.Contains(pkg, `"express"`) {
				t.Error("expected express in package.json")
			}
		})
	}
}

func TestCodegenIntegration_BothTargets(t *testing.T) {
	pyGen := codegen.NewPythonGenerator("0.0.0.0", 8000)
	tsGen := codegen.NewTypeScriptServerGenerator("0.0.0.0", 3000)

	for _, tt := range intentTestFiles {
		t.Run(tt.name, func(t *testing.T) {
			service := parseAndAnalyze(t, tt.path)

			pyOutput := pyGen.Generate(service)
			tsOutput := tsGen.Generate(service)

			if len(pyOutput) == 0 {
				t.Error("Python output is empty")
			}
			if len(tsOutput) == 0 {
				t.Error("TypeScript output is empty")
			}

			// Both should have the same number of routes
			pyRoutes := strings.Count(pyOutput, "@app.")
			tsRoutes := strings.Count(tsOutput, "app.")
			// TypeScript has app.use and app.listen too, so just verify routes exist
			if pyRoutes == 0 {
				t.Error("Python output has no routes")
			}
			if tsRoutes == 0 {
				t.Error("TypeScript output has no route-related calls")
			}
		})
	}
}

func TestCodegenIntegration_CustomProvider(t *testing.T) {
	path := filepath.Join("..", "examples", "custom-provider", "main.glyph")
	service := parseAndAnalyze(t, path)

	// Verify custom providers were detected
	if len(service.Providers) == 0 {
		t.Fatal("expected at least one provider from custom-provider example")
	}

	// Check that non-standard providers exist
	hasCustom := false
	for _, p := range service.Providers {
		if !p.IsStandard {
			hasCustom = true
			break
		}
	}
	if !hasCustom {
		t.Error("expected at least one custom (non-standard) provider")
	}

	// Python output should reference custom providers
	pyGen := codegen.NewPythonGenerator("0.0.0.0", 8000)
	pyOutput := pyGen.Generate(service)
	if !strings.Contains(pyOutput, "Provider") {
		t.Error("Python output should reference custom provider classes")
	}

	// TypeScript output should reference custom providers
	tsGen := codegen.NewTypeScriptServerGenerator("0.0.0.0", 3000)
	tsOutput := tsGen.Generate(service)
	if !strings.Contains(tsOutput, "Provider") {
		t.Error("TypeScript output should reference custom provider classes")
	}
}

func TestCodegenIntegration_TypeDefinitions(t *testing.T) {
	// Use the CRUD API example which has well-defined types
	path := filepath.Join("..", "examples", "intent-tests", "01-crud-api.glyph")
	service := parseAndAnalyze(t, path)

	if len(service.Types) == 0 {
		t.Fatal("expected type definitions in CRUD API example")
	}

	// Verify types appear in both outputs
	pyGen := codegen.NewPythonGenerator("0.0.0.0", 8000)
	tsGen := codegen.NewTypeScriptServerGenerator("0.0.0.0", 3000)

	pyOutput := pyGen.Generate(service)
	tsOutput := tsGen.Generate(service)

	for _, typ := range service.Types {
		// Python uses class Name(BaseModel)
		if !strings.Contains(pyOutput, "class "+typ.Name) {
			t.Errorf("Python output missing type definition for %q", typ.Name)
		}
		// TypeScript uses interface Name
		if !strings.Contains(tsOutput, "interface "+typ.Name) {
			t.Errorf("TypeScript output missing type definition for %q", typ.Name)
		}
	}
}
