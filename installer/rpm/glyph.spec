Name:           glyph
Version:        1.0.0
Release:        1%{?dist}
Summary:        AI-First Backend Language

License:        MIT
URL:            https://github.com/glyphlang/glyph
Source0:        glyph-linux-amd64

BuildArch:      x86_64
ExclusiveArch:  x86_64

%description
GlyphLang is a programming language designed for AI/LLM code generation.
Features symbol-based syntax, sub-microsecond compilation, and built-in
security scanning. Perfect for building REST APIs, WebSocket servers,
and microservices.

Key features:
- Symbol-based syntax (@, $, %, >, +, :) optimized for LLMs
- 867ns compilation time
- 2.93 ns/op VM execution
- Built-in SQL injection and XSS detection
- WebSocket support with room-based messaging

%install
mkdir -p %{buildroot}%{_bindir}
install -m 755 %{SOURCE0} %{buildroot}%{_bindir}/glyph

%files
%{_bindir}/glyph

%changelog
* Sun Dec 15 2025 GlyphLang Team <hello@glyph-lang.dev> - 1.0.0-1
- Production release v1.0.0
- 640+ tests passing
- Sub-microsecond compilation (867ns)
- WebSocket support with room-based messaging
- Built-in security scanning (SQL injection, XSS)
- JIT compilation with type specialization
- Enhanced error messages with suggestions
- Production-ready VM execution
