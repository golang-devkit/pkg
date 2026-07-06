---
name: markdown-language-fix
description: "Proofread and correct Markdown content for grammar, wording, and mixed-language consistency while preserving meaning and Markdown structure."
argument-hint: "Files or folders to process, optional target language (example: docs/**/*.md --lang=en)"
agent: agent
---

# Markdown Language Quality Fixer

## Task
Review Markdown files and correct language-quality problems, including:
- Grammar errors
- Incorrect word usage
- Awkward phrasing
- Mixed-language inconsistencies

Preserve the original meaning, technical intent, and Markdown formatting.

## Inputs
- Required: file path(s), folder(s), or glob patterns to process
- Optional: target language flag (for example `--lang=en` or `--lang=vi`)

If target language is not specified:
1. Use English as the default target language.
2. Translate full prose content to English.
3. Keep required literals and technical tokens unchanged.

## Rules
1. Keep all Markdown structure unchanged unless a fix is required for readability:
   - Preserve heading levels and order.
   - Preserve lists, tables, code fences, links, and blockquotes.
2. Never alter code semantics:
   - Do not modify code inside fenced code blocks unless the user explicitly asks.
   - Do not rename API fields, function names, environment variables, or commands.
3. Improve language quality with minimal semantic drift:
   - Fix grammar, agreement, tense, punctuation, and capitalization.
   - Replace incorrect or unnatural word choices with domain-appropriate wording.
   - Rewrite only as much as needed for clarity and correctness.
   - Prefer minimal edits over stylistic rewrites.
4. Handle multilingual content carefully:
   - Keep proper nouns, product names, and technical tokens as-is.
   - Translate all prose to the target language.
   - If no target language is provided, standardize to English.
5. Preserve document intent and voice:
   - Keep instructional, formal, or concise style consistent with the source.
   - Avoid adding new claims or removing important caveats.

## Process
1. Identify candidate files from the provided paths/patterns.
2. For each file, scan prose sections for grammar, wording, and language-mixing issues.
3. Apply safe edits directly in the file.
4. Re-read changed sections to ensure Markdown validity and semantic preservation.

## Output
Provide:
- Files updated
- Per-file summary of major corrections (grammar, wording, language consistency)
- Any ambiguous passages that may need user confirmation

If no changes are needed, explicitly state that the file(s) already meet the quality rules.
