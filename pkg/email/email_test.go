package email

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient() (*Client, *MockProvider) {
	mock := NewMockProvider("test")
	client := NewClient(mock)
	return client, mock
}

func TestNewClient(t *testing.T) {
	client, _ := newTestClient()
	assert.NotNil(t, client)
	assert.Equal(t, "test", client.ProviderName())
	assert.Empty(t, client.GetTemplates())
}

func TestSendBasicEmail(t *testing.T) {
	client, mock := newTestClient()
	err := client.Send(&Message{
		To:      []string{"user@example.com"},
		Subject: "Hello",
		Body:    "World",
	})
	require.NoError(t, err)
	require.Len(t, mock.Sent, 1)
	assert.Equal(t, []string{"user@example.com"}, mock.Sent[0].To)
	assert.Equal(t, "Hello", mock.Sent[0].Subject)
	assert.Equal(t, "World", mock.Sent[0].Body)
}

func TestSendHTMLEmail(t *testing.T) {
	client, mock := newTestClient()
	err := client.Send(&Message{
		To:      []string{"user@example.com"},
		Subject: "Welcome",
		HTML:    "<h1>Welcome!</h1>",
	})
	require.NoError(t, err)
	require.Len(t, mock.Sent, 1)
	assert.Equal(t, "<h1>Welcome!</h1>", mock.Sent[0].HTML)
}

func TestSendWithCCAndBCC(t *testing.T) {
	client, mock := newTestClient()
	err := client.Send(&Message{
		To:      []string{"to@example.com"},
		CC:      []string{"cc@example.com"},
		BCC:     []string{"bcc@example.com"},
		Subject: "Test",
		Body:    "Content",
	})
	require.NoError(t, err)
	require.Len(t, mock.Sent, 1)
	assert.Equal(t, []string{"cc@example.com"}, mock.Sent[0].CC)
	assert.Equal(t, []string{"bcc@example.com"}, mock.Sent[0].BCC)
}

func TestSendWithAttachments(t *testing.T) {
	client, mock := newTestClient()
	err := client.Send(&Message{
		To:      []string{"user@example.com"},
		Subject: "Invoice",
		Body:    "See attached.",
		Attachments: []Attachment{
			{Filename: "invoice.pdf", Content: []byte("pdf-data"), ContentType: "application/pdf"},
		},
	})
	require.NoError(t, err)
	require.Len(t, mock.Sent, 1)
	require.Len(t, mock.Sent[0].Attachments, 1)
	assert.Equal(t, "invoice.pdf", mock.Sent[0].Attachments[0].Filename)
	assert.Equal(t, "application/pdf", mock.Sent[0].Attachments[0].ContentType)
}

func TestSendValidationNoRecipient(t *testing.T) {
	client, _ := newTestClient()
	err := client.Send(&Message{
		Subject: "Test",
		Body:    "Content",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one recipient")
}

func TestSendValidationInvalidEmail(t *testing.T) {
	client, _ := newTestClient()
	err := client.Send(&Message{
		To:      []string{"not-an-email"},
		Subject: "Test",
		Body:    "Content",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email address")
}

func TestSendValidationNoSubject(t *testing.T) {
	client, _ := newTestClient()
	err := client.Send(&Message{
		To:   []string{"user@example.com"},
		Body: "Content",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subject is required")
}

func TestSendValidationNoBody(t *testing.T) {
	client, _ := newTestClient()
	err := client.Send(&Message{
		To:      []string{"user@example.com"},
		Subject: "Test",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "body or HTML content is required")
}

func TestSendValidationInvalidCC(t *testing.T) {
	client, _ := newTestClient()
	err := client.Send(&Message{
		To:      []string{"user@example.com"},
		CC:      []string{"bad"},
		Subject: "Test",
		Body:    "Content",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CC email")
}

func TestSendValidationInvalidBCC(t *testing.T) {
	client, _ := newTestClient()
	err := client.Send(&Message{
		To:      []string{"user@example.com"},
		BCC:     []string{"bad"},
		Subject: "Test",
		Body:    "Content",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid BCC email")
}

func TestProviderError(t *testing.T) {
	client, mock := newTestClient()
	mock.SendErr = fmt.Errorf("provider unavailable")
	err := client.Send(&Message{
		To:      []string{"user@example.com"},
		Subject: "Test",
		Body:    "Content",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider unavailable")
}

func TestDefaults(t *testing.T) {
	client, mock := newTestClient()
	client.SetDefaults(MessageDefaults{
		From:    "noreply@company.com",
		ReplyTo: "support@company.com",
		Headers: map[string]string{"X-App": "glyph"},
	})
	defaults := client.GetDefaults()
	assert.Equal(t, "noreply@company.com", defaults.From)

	err := client.Send(&Message{
		To:      []string{"user@example.com"},
		Subject: "Test",
		Body:    "Content",
	})
	require.NoError(t, err)
	require.Len(t, mock.Sent, 1)
	assert.Equal(t, "noreply@company.com", mock.Sent[0].From)
	assert.Equal(t, "support@company.com", mock.Sent[0].ReplyTo)
	assert.Equal(t, "glyph", mock.Sent[0].Headers["X-App"])
}

func TestDefaultsDoNotOverrideExplicit(t *testing.T) {
	client, mock := newTestClient()
	client.SetDefaults(MessageDefaults{
		From: "default@company.com",
	})
	err := client.Send(&Message{
		From:    "custom@company.com",
		To:      []string{"user@example.com"},
		Subject: "Test",
		Body:    "Content",
	})
	require.NoError(t, err)
	assert.Equal(t, "custom@company.com", mock.Sent[0].From)
}

func TestRegisterTemplate(t *testing.T) {
	client, _ := newTestClient()
	err := client.RegisterTemplate(&Template{
		Name:    "welcome",
		Subject: "Welcome, {{.Name}}!",
		HTML:    "<h1>Hello {{.Name}}</h1>",
	})
	require.NoError(t, err)
	tmpl, ok := client.GetTemplate("welcome")
	assert.True(t, ok)
	assert.Equal(t, "welcome", tmpl.Name)
}

func TestRegisterTemplateEmptyName(t *testing.T) {
	client, _ := newTestClient()
	err := client.RegisterTemplate(&Template{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template name is required")
}

func TestSendTemplate(t *testing.T) {
	client, mock := newTestClient()
	client.SetDefaults(MessageDefaults{From: "noreply@app.com"})
	err := client.RegisterTemplate(&Template{
		Name:    "welcome",
		Subject: "Welcome, {{.Name}}!",
		HTML:    "<h1>Hello {{.Name}}, your ID is {{.ID}}</h1>",
	})
	require.NoError(t, err)

	err = client.SendTemplate(
		[]string{"user@example.com"},
		"welcome",
		map[string]interface{}{"Name": "Alice", "ID": 42},
	)
	require.NoError(t, err)
	require.Len(t, mock.Sent, 1)
	assert.Equal(t, "Welcome, Alice!", mock.Sent[0].Subject)
	assert.Contains(t, mock.Sent[0].HTML, "Hello Alice")
	assert.Contains(t, mock.Sent[0].HTML, "42")
}

func TestSendTemplateNotFound(t *testing.T) {
	client, _ := newTestClient()
	client.SetDefaults(MessageDefaults{From: "noreply@app.com"})
	err := client.SendTemplate(
		[]string{"user@example.com"},
		"nonexistent",
		nil,
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSendTemplatePlainText(t *testing.T) {
	client, mock := newTestClient()
	client.SetDefaults(MessageDefaults{From: "noreply@app.com"})
	err := client.RegisterTemplate(&Template{
		Name:    "reset",
		Subject: "Password Reset",
		Body:    "Hi {{.Name}}, click here to reset: {{.URL}}",
	})
	require.NoError(t, err)

	err = client.SendTemplate(
		[]string{"user@example.com"},
		"reset",
		map[string]interface{}{"Name": "Bob", "URL": "https://example.com/reset/abc"},
	)
	require.NoError(t, err)
	require.Len(t, mock.Sent, 1)
	assert.Equal(t, "Password Reset", mock.Sent[0].Subject)
	assert.Contains(t, mock.Sent[0].Body, "Hi Bob")
	assert.Contains(t, mock.Sent[0].Body, "https://example.com/reset/abc")
}

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		addr  string
		valid bool
	}{
		{"user@example.com", true},
		{"name@sub.domain.com", true},
		{"", false},
		{"noatsign", false},
		{"@nodomain", false},
		{"nouser@", false},
		{"user@nodot", false},
	}
	for _, tc := range tests {
		t.Run(tc.addr, func(t *testing.T) {
			assert.Equal(t, tc.valid, isValidEmail(tc.addr))
		})
	}
}

func TestMultipleRecipients(t *testing.T) {
	client, mock := newTestClient()
	err := client.Send(&Message{
		To:      []string{"a@example.com", "b@example.com", "c@example.com"},
		Subject: "Broadcast",
		Body:    "Hello all",
	})
	require.NoError(t, err)
	require.Len(t, mock.Sent, 1)
	assert.Len(t, mock.Sent[0].To, 3)
}

func TestInlineAttachment(t *testing.T) {
	client, mock := newTestClient()
	err := client.Send(&Message{
		To:      []string{"user@example.com"},
		Subject: "Logo",
		HTML:    "<img src='cid:logo'>",
		Attachments: []Attachment{
			{
				Filename:    "logo.png",
				Content:     []byte("png-data"),
				ContentType: "image/png",
				Inline:      true,
				ContentID:   "logo",
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, mock.Sent[0].Attachments, 1)
	assert.True(t, mock.Sent[0].Attachments[0].Inline)
	assert.Equal(t, "logo", mock.Sent[0].Attachments[0].ContentID)
}

func TestMessageTags(t *testing.T) {
	client, mock := newTestClient()
	err := client.Send(&Message{
		To:      []string{"user@example.com"},
		Subject: "Tagged",
		Body:    "Content",
		Tags:    []string{"transactional", "welcome"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"transactional", "welcome"}, mock.Sent[0].Tags)
}

func TestGetTemplates(t *testing.T) {
	client, _ := newTestClient()
	_ = client.RegisterTemplate(&Template{Name: "a", Subject: "A", Body: "a"})
	_ = client.RegisterTemplate(&Template{Name: "b", Subject: "B", Body: "b"})
	templates := client.GetTemplates()
	assert.Len(t, templates, 2)
	assert.NotNil(t, templates["a"])
	assert.NotNil(t, templates["b"])
}

func TestProviderConfigs(t *testing.T) {
	// Verify config structs are usable
	_ = SMTPConfig{Host: "smtp.example.com", Port: 587, Username: "user", Password: "pass", UseTLS: true}
	_ = SendGridConfig{APIKey: "sg-key"}
	_ = SESConfig{Region: "us-east-1", AccessKeyID: "id", SecretAccessKey: "secret"}
	_ = MailgunConfig{Domain: "mg.example.com", APIKey: "mg-key"}
	_ = ResendConfig{APIKey: "re-key"}
	_ = PostmarkConfig{ServerToken: "pm-token"}
}

func TestMockProvider(t *testing.T) {
	mock := NewMockProvider("sendgrid")
	assert.Equal(t, "sendgrid", mock.Name())
	err := mock.Send(&Message{To: []string{"a@b.com"}, Subject: "S", Body: "B"})
	require.NoError(t, err)
	assert.Len(t, mock.Sent, 1)
}
