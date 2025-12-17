# Homebrew Formula for GlyphLang
# To install: brew tap glyph-lang/tap && brew install glyph
# Or: brew install glyph-lang/tap/glyph

class Glyph < Formula
  desc "AI-First Backend Language - Symbol-based syntax for LLM code generation"
  homepage "https://github.com/glyphlang/glyph"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/glyphlang/glyph/releases/download/v1.0.0/glyph-darwin-arm64.zip"
      sha256 "PLACEHOLDER_ARM64_SHA256"
    else
      url "https://github.com/glyphlang/glyph/releases/download/v1.0.0/glyph-darwin-amd64.zip"
      sha256 "PLACEHOLDER_AMD64_SHA256"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/glyphlang/glyph/releases/download/v1.0.0/glyph-linux-arm64.zip"
      sha256 "PLACEHOLDER_LINUX_ARM64_SHA256"
    else
      url "https://github.com/glyphlang/glyph/releases/download/v1.0.0/glyph-linux-amd64.zip"
      sha256 "PLACEHOLDER_LINUX_AMD64_SHA256"
    end
  end

  def install
    binary_name = if OS.mac?
      Hardware::CPU.arm? ? "glyph-darwin-arm64" : "glyph-darwin-amd64"
    else
      Hardware::CPU.arm? ? "glyph-linux-arm64" : "glyph-linux-amd64"
    end

    # The zip contains the binary directly
    if File.exist?(binary_name)
      bin.install binary_name => "glyph"
    elsif File.exist?("glyph")
      bin.install "glyph"
    else
      # Find any glyph binary
      glyph_binary = Dir["glyph*"].reject { |f| f.end_with?(".zip") }.first
      bin.install glyph_binary => "glyph" if glyph_binary
    end
  end

  test do
    assert_match "glyph version", shell_output("#{bin}/glyph --version")
  end
end
