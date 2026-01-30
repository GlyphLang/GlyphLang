package oauth2

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// GeneratePKCE generates PKCE (Proof Key for Code Exchange) parameters.
// Uses the S256 method as recommended by RFC 7636.
func GeneratePKCE() (*PKCEParams, error) {
	// Generate a random code verifier (43-128 characters per RFC 7636)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, err
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Generate code challenge using SHA-256
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return &PKCEParams{
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
		Method:        "S256",
	}, nil
}

// VerifyPKCE verifies a PKCE code verifier against a code challenge
func VerifyPKCE(codeVerifier, codeChallenge string) bool {
	hash := sha256.Sum256([]byte(codeVerifier))
	expected := base64.RawURLEncoding.EncodeToString(hash[:])
	return expected == codeChallenge
}
