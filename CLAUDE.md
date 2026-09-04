# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`github.com/shouni/go-character-kit` is a small **library module** (no `main` package, no binary)
that loads, validates and looks up character definitions — `id`, `name`, `seed`, `reference_url` /
`reference_urls`, `visual_cues`, `is_default` — for image and comic generation pipelines.

It does nothing else. It builds no prompts, calls no AI, touches no network or filesystem beyond
the one embedded JSON in `assets`. The module has **zero dependencies**: `go.mod` lists nothing
but the `go` directive, and the tests use the standard library alone. Keep it that way — a
`require` here reaches every consumer, and a definition-holder is the last place that should
drag one in.

Which repositories read these definitions is a question for the sibling `go.mod`s, not for this
file — a list written here goes stale and some of those repos are private:

    grep -l "shouni/go-character-kit" ~/GolandProjects/*/go.mod

## Commands

```bash
go build ./...
go vet ./...
go test ./...
go test ./character -run TestWithSeedOverride -v   # single test
test -z "$(gofmt -l .)"                           # CI fails on unformatted files
golangci-lint run                                 # config in .golangci.yml
```

`.github/workflows/ci.yml` is a thin caller of the shared
`shouni/workflows/.github/workflows/go-ci.yml@v1` and passes no inputs, so it gets that
workflow's defaults (build / vet / gofmt / race test, golangci-lint, govulncheck) on `main` and
`develop`. The job list and the linter pin live in that shared workflow, not here.

## Packages

Two packages, at the repository root — no `pkg/` prefix.

- **`character`** — the data model (`Character`), the validated collection (`Characters`), JSON
  parsing, and the lookup helpers. Depends on nothing.
- **`assets`** — one function, `LoadCharacters()`, over a `go:embed`-ed
  `assets/characters/characters.json`. It is a separate package **because the embedded roster is a
  convenience, not part of the model**: a consumer supplying its own JSON imports `character` only
  and never carries the bundled bytes in its binary. Everything `assets` knows how to do is
  `character.ParseCharacters` on a fixed byte slice.

## Design decisions

- **Validated at construction, read-only afterwards.** `NewCharacters` / `ParseCharacters` are the
  only doors in, and both run `validateList` before anything is stored. The input list is deep-copied
  (`cloneList`), and every accessor — `All`, `GetCharacter`, `GetDefault`, `GetCharacterWithDefault`
  — hands back a copy too. There is deliberately no path by which a character definition can change
  while a generation run is in flight: the whole point of the type is that the seed and the
  reference URL a pipeline read at panel 1 are the same ones it reads at panel 12. `WithSeedOverride`
  follows the same rule — it derives a **new** collection through `newValidated` rather than mutating
  the receiver, and returns the receiver unchanged when the ID does not match.
- **"Not found" is decided by the type, not by every caller.** `GetCharacterWithDefault` falls back
  to the `is_default` character for an unknown or empty ID, so consumers do not each write a nil
  check plus their own answer to "then who?". `GetCharacter` still returns `nil` for callers that
  genuinely need to know. Validation enforces **at most one** `is_default`, which is what makes the
  fallback a single well-defined answer; more than one is rejected with all the offending IDs
  listed. Every method is also nil-receiver safe (`Len` → 0, `All`/`GetCharacter` → nil), so an
  optional roster needs no wrapper.
- **Reference images are chosen by aspect ratio.** `Character.ReferenceURLFor(aspectRatio)` prefers
  the `reference_urls` entry matching the ratio being generated and falls back to the
  aspect-agnostic `reference_url`. The same character needs a different reference depending on the
  target shape — feeding a wide three-pose sheet into a portrait keyframe is what makes colours,
  props and hair drift between generations. The rationale sits on the `ReferenceURLs` field's doc
  comment; keep it there when the field changes.
- **Ratio keys are validated for shape, not for membership.** `aspectRatioKeyPattern` accepts any
  `\d+:\d+`, so this module has no opinion on which ratios exist — that vocabulary belongs to the
  generation kits. The check exists only to catch the typo (`"16x9"`) that would otherwise fail
  silently as a lookup miss that quietly falls back to `reference_url`.
- **ID matching is case-insensitive** (`byID` is keyed on the lower-cased ID, and duplicates that
  differ only in case are rejected at construction). IDs arrive from AI output and hand-written
  JSON alike, so a case difference should not decide who gets drawn.
- **`Seed` is `*int64`, not a bare `int64`.** Nil means "this character pins no seed", which the
  consuming kit's seed-resolution chain must be able to tell apart from a real seed of 0.

## Conventions

- Doc comments, error messages and validation text are in **Japanese**, in です／ます style.
- Tests are standard-library only (`testing` + hand-written comparisons), table-driven where the
  cases justify it, and named for the behaviour they pin (`TestAccessorsReturnCopies`,
  `TestNilReceiverSafety`). The copy-on-read and nil-receiver guarantees above are load-bearing for
  the consumers, so they each have a dedicated test — do not fold them into a broader case.
- README.md is the user-facing entry point (definition format, validation list, usage). Keep the
  field table and the validation list in sync when changing either.
