#!/usr/bin/env python3
"""Generation Accuracy Benchmark - measure first-try validity of LLM-generated
GlyphLang versus FastAPI code for equivalent tasks.

This tests the core AI-first claim: can a model that has never seen GlyphLang
in training produce valid programs from the notation spec alone, at a rate
comparable to FastAPI (which is heavily represented in training data)?

Conditions:
  - glyph:   the model receives docs/GLYPH_NOTATION_SPEC.md as context
             (realistic usage - the spec is small enough to always include).
             Output is checked with `glyph validate` (syntax + semantics).
  - fastapi: no special context (realistic baseline for a well-known
             framework). Output is checked with ast.parse (syntax only, so
             this is a lenient bar - noted in results).

Refusals and API errors are recorded as their own outcomes, not failures.
No model fallback is configured on purpose: the eval must measure the named
model, not a silent substitute.

Requirements:
    pip install anthropic
    ANTHROPIC_API_KEY set (or an `ant auth login` profile)
    Go toolchain (to build the glyph binary if not on PATH)

Usage:
    python benchmarks/bench_generation_accuracy.py [--model claude-opus-5] [--trials 1]
"""

import argparse
import ast
import json
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

try:
    from anthropic import Anthropic
except ImportError:
    sys.exit("The anthropic SDK is required. Install with: pip install anthropic")

REPO_ROOT = Path(__file__).resolve().parent.parent
SPEC_PATH = REPO_ROOT / "docs" / "GLYPH_NOTATION_SPEC.md"

# Task descriptions are language-neutral; the condition supplies the framing.
TASKS = [
    ("hello_world",
     'an HTTP API with a GET /hello endpoint that returns the JSON object {"message": "Hello, World!"}'),
    ("path_param",
     "an HTTP API with a User type (id: required int, name: required string, email: required string) "
     "and a GET /users/:id endpoint that returns a User"),
    ("post_validation",
     "an HTTP API with a POST /users endpoint that accepts a name (string, 1-100 chars) and email "
     "(valid email address), validates the input, and returns the created user"),
    ("protected_route",
     "an HTTP API with a GET /api/data endpoint protected by JWT authentication and rate limited to "
     '100 requests per minute, returning {"data": "secret"}'),
    ("crud",
     "an HTTP API for a Todo type with id (required int) and title (required string), with three "
     "endpoints: GET /todos returning all todos, POST /todos creating one, and DELETE /todos/:id "
     "deleting one and returning {\"success\": true}"),
]


def build_glyph_system() -> str:
    spec = SPEC_PATH.read_text(encoding="utf-8")
    return (
        "You write GlyphLang programs. GlyphLang is a compact, AI-first backend language. "
        "Its full notation specification follows.\n\n" + spec +
        "\n\nRespond with only the GlyphLang source code. No markdown fences, no explanation."
    )


FASTAPI_SYSTEM = (
    "You write Python FastAPI applications. "
    "Respond with only the Python source code. No markdown fences, no explanation."
)


def find_glyph_binary() -> str:
    exe = shutil.which("glyph")
    if exe:
        return exe
    out = Path(tempfile.mkdtemp(prefix="glyph_eval_")) / ("glyph.exe" if os.name == "nt" else "glyph")
    subprocess.run(["go", "build", "-o", str(out), "./cmd/glyph"], cwd=REPO_ROOT, check=True)
    return str(out)


def extract_code(text: str) -> str:
    # Strip markdown fences if the model ignores the no-fences instruction.
    t = text.strip()
    if t.startswith("```"):
        lines = t.splitlines()[1:]
        if lines and lines[-1].strip().startswith("```"):
            lines = lines[:-1]
        t = "\n".join(lines).strip()
    return t


def validate_glyph(code: str, glyph_bin: str) -> tuple[bool, str]:
    with tempfile.NamedTemporaryFile("w", suffix=".glyph", delete=False, encoding="utf-8") as f:
        f.write(code)
        path = f.name
    try:
        result = subprocess.run([glyph_bin, "validate", path],
                                capture_output=True, text=True, timeout=30)
        detail = (result.stdout + result.stderr).strip().splitlines()
        return result.returncode == 0, detail[-1] if detail else ""
    finally:
        os.unlink(path)


def validate_python(code: str) -> tuple[bool, str]:
    try:
        ast.parse(code)
        return True, ""
    except SyntaxError as e:
        return False, str(e)


def run_trial(client: Anthropic, model: str, system: str, prompt: str) -> tuple[str, str]:
    """Returns (status, payload): status is 'ok', 'refusal', or 'error'."""
    try:
        resp = client.messages.create(
            model=model,
            max_tokens=16000,
            system=system,
            messages=[{"role": "user", "content": prompt}],
        )
    except Exception as e:
        return "error", f"{type(e).__name__}: {e}"
    if resp.stop_reason == "refusal":
        return "refusal", ""
    text = "".join(b.text for b in resp.content if b.type == "text")
    return "ok", extract_code(text)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--model", default="claude-opus-5")
    parser.add_argument("--trials", type=int, default=1)
    parser.add_argument("--tasks", nargs="*", help="Subset of task names to run")
    args = parser.parse_args()

    glyph_bin = find_glyph_binary()
    client = Anthropic()
    glyph_system = build_glyph_system()

    tasks = [t for t in TASKS if not args.tasks or t[0] in args.tasks]
    results = []

    for name, desc in tasks:
        for condition, system, phrasing in (
            ("glyph", glyph_system, f"Write {desc} in GlyphLang."),
            ("fastapi", FASTAPI_SYSTEM, f"Write {desc} using FastAPI."),
        ):
            for trial in range(args.trials):
                status, payload = run_trial(client, args.model, system, phrasing)
                if status == "ok":
                    if condition == "glyph":
                        passed, detail = validate_glyph(payload, glyph_bin)
                    else:
                        passed, detail = validate_python(payload)
                    outcome = "pass" if passed else "fail"
                else:
                    outcome, detail = status, payload
                results.append({"task": name, "condition": condition, "trial": trial,
                                "outcome": outcome, "detail": detail})
                print(f"{name:16s} {condition:8s} trial {trial}: {outcome}"
                      + (f"  ({detail})" if outcome != "pass" and detail else ""))

    summary = {}
    for condition in ("glyph", "fastapi"):
        rows = [r for r in results if r["condition"] == condition]
        passes = sum(1 for r in rows if r["outcome"] == "pass")
        summary[condition] = {
            "trials": len(rows),
            "passes": passes,
            "refusals": sum(1 for r in rows if r["outcome"] == "refusal"),
            "errors": sum(1 for r in rows if r["outcome"] == "error"),
            "pass_rate": round(passes / len(rows), 3) if rows else None,
        }

    print()
    print(json.dumps({
        "model": args.model,
        "note": "glyph checked with `glyph validate` (syntax+semantics); "
                "fastapi checked with ast.parse (syntax only, lenient bar)",
        "summary": summary,
    }, indent=2))


if __name__ == "__main__":
    main()
