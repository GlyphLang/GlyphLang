package interpreter

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TemplateEngine handles HTML template rendering with caching
type TemplateEngine struct {
	baseDir   string
	cache     map[string]*cachedTemplate
	cacheLock sync.RWMutex
	funcMap   template.FuncMap
}

type cachedTemplate struct {
	tmpl     *template.Template
	modTime  time.Time
	filePath string
}

// NewTemplateEngine creates a new template engine with the given base directory
func NewTemplateEngine(baseDir string) *TemplateEngine {
	return &TemplateEngine{
		baseDir: baseDir,
		cache:   make(map[string]*cachedTemplate),
		funcMap: defaultFuncMap(),
	}
}

// defaultFuncMap returns the default template functions
func defaultFuncMap() template.FuncMap {
	return template.FuncMap{
		// Date/time formatting
		"formatDate": func(t time.Time, layout string) string {
			return t.Format(layout)
		},
		"now": func() time.Time {
			return time.Now()
		},

		// String helpers
		"upper": func(s string) string {
			return s
		},
		"lower": func(s string) string {
			return s
		},

		// Comparison helpers
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"ne": func(a, b interface{}) bool {
			return a != b
		},

		// Safe HTML (already escaped)
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},

		// JSON encoding helper
		"json": func(v interface{}) string {
			// Simple JSON-like representation
			return fmt.Sprintf("%v", v)
		},

		// Default value helper
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
	}
}

// SetBaseDir sets the base directory for template files
func (e *TemplateEngine) SetBaseDir(dir string) {
	e.baseDir = dir
}

// AddFunc adds a custom function to the template engine
func (e *TemplateEngine) AddFunc(name string, fn interface{}) {
	e.funcMap[name] = fn
}

// Render renders a template with the given data
func (e *TemplateEngine) Render(templatePath string, data interface{}) (string, error) {
	// Resolve template path
	fullPath := templatePath
	if !filepath.IsAbs(templatePath) && e.baseDir != "" {
		fullPath = filepath.Join(e.baseDir, templatePath)
	}

	// Clean the path
	fullPath = filepath.Clean(fullPath)

	// Get or load template
	tmpl, err := e.getTemplate(fullPath)
	if err != nil {
		return "", err
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// getTemplate gets a template from cache or loads it from disk
func (e *TemplateEngine) getTemplate(fullPath string) (*template.Template, error) {
	// Check file exists and get mod time
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template not found: %s", fullPath)
		}
		return nil, fmt.Errorf("failed to stat template: %w", err)
	}

	// Check cache
	e.cacheLock.RLock()
	cached, exists := e.cache[fullPath]
	e.cacheLock.RUnlock()

	if exists && cached.modTime.Equal(info.ModTime()) {
		return cached.tmpl, nil
	}

	// Load template
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New(filepath.Base(fullPath)).Funcs(e.funcMap).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Update cache
	e.cacheLock.Lock()
	e.cache[fullPath] = &cachedTemplate{
		tmpl:     tmpl,
		modTime:  info.ModTime(),
		filePath: fullPath,
	}
	e.cacheLock.Unlock()

	return tmpl, nil
}

// ClearCache clears the template cache
func (e *TemplateEngine) ClearCache() {
	e.cacheLock.Lock()
	e.cache = make(map[string]*cachedTemplate)
	e.cacheLock.Unlock()
}

// RenderTemplate is a method on Interpreter to render templates
func (i *Interpreter) RenderTemplate(templatePath string, data interface{}) (string, error) {
	if i.templateEngine == nil {
		// Create default template engine if not set
		i.templateEngine = NewTemplateEngine("")
	}
	return i.templateEngine.Render(templatePath, data)
}

// SetTemplateDir sets the base directory for templates
func (i *Interpreter) SetTemplateDir(dir string) {
	if i.templateEngine == nil {
		i.templateEngine = NewTemplateEngine(dir)
	} else {
		i.templateEngine.SetBaseDir(dir)
	}
}
