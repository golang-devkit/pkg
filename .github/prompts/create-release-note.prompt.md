---
name: release-note
description: "Generate or update release notes in CHANGELOG.md for a given git tag or commit range. Use when releasing a new version, backfilling missing changelog entries, or fixing markdown style issues in CHANGELOG.md."
argument-hint: "Target tag or commit (e.g. v1.1.115), or leave blank to be prompted."
agent: agent
model: Claude Haiku 4.5
tools: [execute/runInTerminal, read/readFile, edit/editFiles]
---

# Release Note Generator

You are generating or updating release notes for this Go model repository.

## Step 1 — Determine the target tag / commit range

If the user did not supply a tag or commit in their message, ask:

> Which tag (or commit SHA) should this release note cover?
> (e.g. `v1.1.115`, or a range like `v1.1.113...v1.1.115`)

Once you have the target, resolve:
- **`$NEW_TAG`** — the tag/commit being released (or the newest end of the range).
- **`$PREV_TAG`** — the immediately preceding tag, obtained by running:

```bash
git describe --tags --abbrev=0 "$NEW_TAG"^
```

If `$NEW_TAG` is not yet a real tag (i.e. it is a planned version), treat `HEAD` as the tip and the most recent existing tag as `$PREV_TAG`.

## Step 2 — Inspect changes with git

Run the following commands to gather raw material for the release note.

```bash
# List commits in range (one-line summary)
git log --oneline "$PREV_TAG".."$NEW_TAG"

# Show full diff limited to Go source files
git diff "$PREV_TAG" "$NEW_TAG" -- '*.go'

# Show only the file-level change summary
git diff --stat "$PREV_TAG" "$NEW_TAG" -- '*.go'
```

Analyse the output:
- Group changed `.go` files by their top-level package directory (e.g. `movil/`, `hr/`, `bkoFx/`).
- Identify struct additions, field additions/removals, new enums, new validate logic, new files, and deleted files.
- Classify each change as **Enhancement**, **Bug Fix**, **Breaking Change**, or **Refactor**.

## Step 3 — Draft the changelog entry

Write a new `## [vX.Y.Z] - YYYY-MM-DD` section following the exact style already used in [CHANGELOG.md](../../CHANGELOG.md):

```markdown
## [vX.Y.Z] - YYYY-MM-DD

**Full Changelog**: [vX.Y.(Z-1)...vX.Y.Z](https://github.com/{{organization}}/{{repository}}/compare/vX.Y.(Z-1)...vX.Y.Z)

### <emoji> <Package> Model — <Short Description> (<Classification>)

#### 📝 Modified Struct — <StructName>

##### <One-line summary of the change>

- **Modified Struct: <StructName>:**
  - Added `<Field>` (<Type>) — <description>
  - Removed `<Field>` — <reason>

#### 🎯 Implementation Details — vX.Y.Z (<Package> Model — <StructName>)

**Files Modified (N file(s)):**

| File | Change | Type | Impact |
|------|--------|------|--------|
| `<path/file.go>` | <what changed> | <Enhancement/Bug Fix/…> | <Breaking/Non-breaking — short impact note> |

#### ✅ Implementation Benefits — vX.Y.Z

- **<Benefit 1>**: <explanation>
- **<Benefit 2>**: <explanation>

---
```

Rules for the draft:
- Use the same emoji convention as existing entries (📱 mobile, 👥 HR, 💱 FX, etc.). Pick the closest match.
- The `Full Changelog` URL must use the real GitHub compare path.
- Keep the table columns aligned.
- End each top-level section with `---`.

## Step 4 — Insert into CHANGELOG.md

Read `CHANGELOG.md` and prepend the new entry **directly after the `# Changelog` heading** (line 1), keeping all existing entries below it unchanged.

Entries must remain sorted **descending by version** (newest first).

If a version block for `$NEW_TAG` already exists, update it in place rather than duplicating it.

## Step 5 — Fix markdown style issues

While editing CHANGELOG.md, also repair any of the following style issues found anywhere in the file:

| Issue | Fix |
|-------|-----|
| Missing blank line before/after a heading | Add blank line |
| Inconsistent heading levels (skipped H4→H6) | Normalize hierarchy |
| Trailing whitespace on a line | Remove it |
| Missing `---` separator between version blocks | Add it |
| Broken or misformatted compare links | Reconstruct from tag names |
| Table columns misaligned (only in new/touched sections) | Re-align |

Do **not** rewrite sections you have not touched unless they contain a clear markdown syntax error.

## Output

- Confirm the version range used: `"Changelog updated: $PREV_TAG → $NEW_TAG"`.
- List each `.go` file that was included in the diff summary.
- Show the exact block that was inserted/updated in CHANGELOG.md.
