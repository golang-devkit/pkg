---
name: markdown-lint
description: 'Check and correct formatting errors in Markdown files using markdownlint. Auto-fixes issues when possible, reports unfixable problems, and lists violations with suggestions.'
argument-hint: 'File patterns or paths to process (e.g., "**/*.md", "docs/", "src/**/*.md")'
agent: agent
---

## Task
Lint and fix Markdown files using markdownlint. Auto-fix formatting errors where possible, report any issues that cannot be automatically fixed, and provide suggestions for violations.

## Inputs
- **File patterns/paths**: Required. Glob patterns or file paths to process (e.g., `**/*.md`, `docs/`, specific files)
- **Scope**: Personal or project-level (optional, defaults to workspace)

## Process
1. Run markdownlint on specified file patterns with auto-fix enabled
2. Capture and parse linting results
3. For each file:
   - Auto-fix violations that markdownlint can resolve
   - Report unfixable issues with rule names and line numbers
   - Provide actionable suggestions for violations
4. Generate summary showing files fixed, issues remaining, and next steps

## Expected Output
- **Fixed files**: List of files successfully corrected
- **Issues requiring manual intervention**: Detailed violations with:
  - File path and line number
  - Rule name and description
  - Specific suggestion or resolution approach
- **Summary**: Count of fixed issues vs. remaining violations
- **Command reference**: Show the exact markdownlint command run for reproducibility

## Notes
- Works with standard Markdown and common flavors (GitHub-Flavored Markdown, CommonMark)
- Can be scoped to specific directories or patterns
- Best used in CI/CD pipelines or pre-commit workflows
