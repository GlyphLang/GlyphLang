// Package email provides email sending capabilities with pluggable providers
// (SendGrid, AWS SES, SMTP, Mailgun, Resend, Postmark) and template support.
package email

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"text/template"
)

// Provider defines the interface that all email providers must implement.
type Provider interface {
	Send(msg *Message) error
	Name() string
}

// Message represents an email message.
type Message struct {
	From        string
	To          []string
	CC          []string
	BCC         []string
	ReplyTo     string
	Subject     string
	Body        string // Plain text body
	HTML        string // HTML body
	Attachments []Attachment
	Headers     map[string]string
	Tags        []string // Provider-specific tags for tracking
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename    string
	Content     []byte
	ContentType string // MIME type (e.g., "application/pdf")
	Inline      bool   // Whether to inline the attachment (for embedded images)
	ContentID   string // Content-ID for inline attachments
}

// Template represents an email template with variable substitution.
type Template struct {
	Name    string
	Subject string // Subject template (supports {{.Var}} syntax)
	Body    string // Plain text body template
	HTML    string // HTML body template
}

// Client is the main email client that manages providers and templates.
type Client struct {
	mu        sync.RWMutex
	provider  Provider
	templates map[string]*Template
	defaults  MessageDefaults
}

// MessageDefaults holds default values applied to all outgoing messages.
type MessageDefaults struct {
	From    string
	ReplyTo string
	Headers map[string]string
}

// NewClient creates a new email client with the given provider.
func NewClient(provider Provider) *Client {
	return &Client{
		provider:  provider,
		templates: make(map[string]*Template),
	}
}

// SetDefaults sets default values for outgoing messages.
func (c *Client) SetDefaults(defaults MessageDefaults) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaults = defaults
}

// GetDefaults returns the current default message values.
func (c *Client) GetDefaults() MessageDefaults {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.defaults
}

// RegisterTemplate registers a named email template.
func (c *Client) RegisterTemplate(tmpl *Template) error {
	if tmpl.Name == "" {
		return fmt.Errorf("template name is required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.templates[tmpl.Name] = tmpl
	return nil
}

// GetTemplate returns a registered template by name.
func (c *Client) GetTemplate(name string) (*Template, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.templates[name]
	return t, ok
}

// GetTemplates returns a copy of all registered templates.
func (c *Client) GetTemplates() map[string]*Template {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]*Template, len(c.templates))
	for k, v := range c.templates {
		result[k] = v
	}
	return result
}

// Send sends an email message through the configured provider.
func (c *Client) Send(msg *Message) error {
	if err := validateMessage(msg); err != nil {
		return err
	}
	c.applyDefaults(msg)
	return c.provider.Send(msg)
}

// SendTemplate sends an email using a registered template with data substitution.
func (c *Client) SendTemplate(to []string, templateName string, data map[string]interface{}) error {
	c.mu.RLock()
	tmpl, ok := c.templates[templateName]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("template %q not found", templateName)
	}

	msg := &Message{To: to}

	if tmpl.Subject != "" {
		rendered, err := renderTemplate(tmpl.Subject, data)
		if err != nil {
			return fmt.Errorf("failed to render subject template: %w", err)
		}
		msg.Subject = rendered
	}

	if tmpl.Body != "" {
		rendered, err := renderTemplate(tmpl.Body, data)
		if err != nil {
			return fmt.Errorf("failed to render body template: %w", err)
		}
		msg.Body = rendered
	}

	if tmpl.HTML != "" {
		rendered, err := renderTemplate(tmpl.HTML, data)
		if err != nil {
			return fmt.Errorf("failed to render HTML template: %w", err)
		}
		msg.HTML = rendered
	}

	return c.Send(msg)
}

// ProviderName returns the name of the configured provider.
func (c *Client) ProviderName() string {
	return c.provider.Name()
}

// applyDefaults applies default values to a message where fields are empty.
func (c *Client) applyDefaults(msg *Message) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if msg.From == "" && c.defaults.From != "" {
		msg.From = c.defaults.From
	}
	if msg.ReplyTo == "" && c.defaults.ReplyTo != "" {
		msg.ReplyTo = c.defaults.ReplyTo
	}
	if c.defaults.Headers != nil {
		if msg.Headers == nil {
			msg.Headers = make(map[string]string)
		}
		for k, v := range c.defaults.Headers {
			if _, exists := msg.Headers[k]; !exists {
				msg.Headers[k] = v
			}
		}
	}
}

// validateMessage validates that a message has required fields.
func validateMessage(msg *Message) error {
	if len(msg.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	for _, addr := range msg.To {
		if !isValidEmail(addr) {
			return fmt.Errorf("invalid email address: %s", addr)
		}
	}
	for _, addr := range msg.CC {
		if !isValidEmail(addr) {
			return fmt.Errorf("invalid CC email address: %s", addr)
		}
	}
	for _, addr := range msg.BCC {
		if !isValidEmail(addr) {
			return fmt.Errorf("invalid BCC email address: %s", addr)
		}
	}
	if msg.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if msg.Body == "" && msg.HTML == "" {
		return fmt.Errorf("body or HTML content is required")
	}
	return nil
}

// isValidEmail performs basic email validation.
func isValidEmail(addr string) bool {
	if addr == "" {
		return false
	}
	parts := strings.SplitN(addr, "@", 2)
	if len(parts) != 2 {
		return false
	}
	if parts[0] == "" || parts[1] == "" {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

// renderTemplate renders a Go text/template string with the given data.
func renderTemplate(tmplStr string, data map[string]interface{}) (string, error) {
	t, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// SMTPConfig holds SMTP connection settings.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
}

// SendGridConfig holds SendGrid provider settings.
type SendGridConfig struct {
	APIKey string
}

// SESConfig holds AWS SES provider settings.
type SESConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

// MailgunConfig holds Mailgun provider settings.
type MailgunConfig struct {
	Domain string
	APIKey string
}

// ResendConfig holds Resend provider settings.
type ResendConfig struct {
	APIKey string
}

// PostmarkConfig holds Postmark provider settings.
type PostmarkConfig struct {
	ServerToken string
}

// MockProvider is a test implementation of the Provider interface.
type MockProvider struct {
	mu      sync.Mutex
	name    string
	Sent    []*Message
	SendErr error
}

// NewMockProvider creates a new mock email provider.
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{name: name}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Send(msg *Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.SendErr != nil {
		return m.SendErr
	}
	m.Sent = append(m.Sent, msg)
	return nil
}
