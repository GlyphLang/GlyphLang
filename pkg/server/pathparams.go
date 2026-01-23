package server

import (
	"net/url"
	"strings"
)

// ExtractRouteParamNames extracts parameter names from a route pattern.
// e.g., "/users/:id/:action" returns ["id", "action"]
// e.g., "/chat/:room" returns ["room"]
func ExtractRouteParamNames(path string) []string {
	params := []string{}
	parts := []rune(path)

	for i := 0; i < len(parts); i++ {
		if parts[i] == ':' {
			// Found a parameter, extract the name
			paramStart := i + 1
			paramEnd := paramStart

			// Find the end of the parameter (next / or end of string)
			for paramEnd < len(parts) && parts[paramEnd] != '/' {
				paramEnd++
			}

			if paramEnd > paramStart {
				paramName := string(parts[paramStart:paramEnd])
				params = append(params, paramName)
			}

			i = paramEnd - 1
		}
	}

	return params
}

// ExtractPathParamValues extracts parameter values from an actual path given a pattern.
// pattern: "/chat/:room" actualPath: "/chat/general" -> {"room": "general"}
// Handles URL-encoded values by decoding them.
func ExtractPathParamValues(pattern, actualPath string) map[string]string {
	params := make(map[string]string)

	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	actualParts := strings.Split(strings.Trim(actualPath, "/"), "/")

	if len(patternParts) != len(actualParts) {
		return params
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			paramName := part[1:]
			// URL-decode the value to handle encoded characters
			value := actualParts[i]
			if decoded, err := url.PathUnescape(value); err == nil {
				value = decoded
			}
			params[paramName] = value
		}
	}

	return params
}

// ConvertPatternToMuxFormat converts Glyph's :param syntax to Go's {param} syntax
// for http.ServeMux pattern matching.
// e.g., "/chat/:room" becomes "/chat/{room}"
func ConvertPatternToMuxFormat(pattern string) string {
	parts := strings.Split(pattern, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}
