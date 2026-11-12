# 🎨 Go Character Kit

[![CI](https://github.com/shouni/go-character-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/shouni/go-character-kit/actions/workflows/ci.yml)
[![Status](https://img.shields.io/badge/Status-Active-brightgreen)](#)
[![Language](https://img.shields.io/badge/Language-Go-blue)](https://go.dev/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shouni/go-character-kit)](https://go.dev/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/shouni/go-character-kit)](https://github.com/shouni/go-character-kit/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/shouni/go-character-kit.svg)](https://pkg.go.dev/github.com/shouni/go-character-kit)

English | [日本語](README.ja.md)

## 🚀 About - Pin the seed and reference assets so characters stay consistent

**Go Character Kit** is a small Go library that safely loads, validates and looks up **character DNA**, defined as JSON, for image and comic generation workflows.

It keeps each character's **seed**, **reference asset URLs**, **visual cues / appearance directions** and the **default character** in one place, in a form the generation pipeline can rely on.

---

## ✨ Features

* **🧬 Character DNA Definition**:
  * Defines `id` / `name` / `seed` / `reference_url` / `reference_urls` / `visual_cues` together — the information needed to keep a character consistent.
* **🔒 Immutable Collection**:
  * `Characters` is validated and copied at construction and is read-only afterwards. The lookup APIs return copies, so callers cannot corrupt its internal state.
* **🔍 Safe Lookup Helpers**:
  * Lookup by ID (case-insensitive), fallback to the default character when no ID is given, per-aspect-ratio reference URL resolution (`ReferenceURLFor`), and derivation of a collection with a replaced seed (`WithSeedOverride`).
* **🛡 Validation First**:
  * Configuration mistakes are caught at parse time (see [Validation](#-validation)).
* **📦 Embedded Character Assets**:
  * The `assets` package loads the character definition JSON bundled with the repository as `go:embed`-ed data.

---

## 🎨 Character Definition

Characters are defined as a JSON array.

```json
[
  {
    "id": "zundamon",
    "name": "Zundamon",
    "seed": 10001,
    "reference_url": "gs://bucket/zundamon.png",
    "reference_urls": {
      "16:9": "gs://bucket/zundamon-16x9.png",
      "9:16": "gs://bucket/zundamon-9x16.png"
    },
    "visual_cues": [
      "vibrant emerald green hair",
      "soybean earmuffs",
      "strictly following the design from reference"
    ],
    "is_default": true
  }
]
```

| Field | Required | Description |
| --- | --- | --- |
| `id` | ✅ | A stable ID identifying the character. Leading or trailing whitespace is not allowed. |
| `name` | ✅ | Display / management name. |
| `visual_cues` | ✅ | Visual traits injected into the generation prompt. |
| `reference_url` | - | Reference image URL for consistency (the default fallback). Accepts `gs://...` and the like. Leave it out for a new character that has no image yet: her first design sheet is drawn from `visual_cues` alone, and `ReferenceURLFor` returns `""`. |
| `reference_urls` | - | Reference image URLs keyed by an aspect-ratio string such as `"16:9"`. `ReferenceURLFor` prefers the entry matching the aspect ratio being generated. |
| `sheets` | - | Supplementary design sheets keyed by kind: `expressions` (the same face in several moods), `outfit` (the costume, front and back), `parts` (accessories laid out singly). They sit beside `reference_url`, the full-body identity anchor, and `SheetURL(kind)` returns `""` for one not drawn yet. Any other key is rejected. |
| `seed` | - | An optional generation seed tied to the character. |
| `is_default` | - | The default character used for fallback. At most one may be set. |

---

## 🚦 Usage

### 1. Load from JSON

```go
package main

import (
	"fmt"
	"log"

	"github.com/shouni/go-character-kit/character"
)

func main() {
	chars, err := character.ParseCharacters([]byte(`[
		{
			"id": "zundamon",
			"name": "Zundamon",
			"reference_url": "gs://bucket/zundamon.png",
			"visual_cues": ["green hair"],
			"seed": 10001,
			"is_default": true
		}
	]`))
	if err != nil {
		log.Fatal(err)
	}

	char := chars.GetCharacterWithDefault("ZUNDAMON")
	fmt.Println(char.ID, char.Name)
}
```

### 2. Initialize from Go structs

```go
chars, err := character.NewCharacters([]character.Character{
	{
		ID:           "zundamon",
		Name:         "Zundamon",
		ReferenceURL: "gs://bucket/zundamon.png",
		VisualCues:   []string{"green hair"},
		IsDefault:    true,
	},
})
```

### 3. Load the bundled character definitions

```go
chars, err := assets.LoadCharacters() // import "github.com/shouni/go-character-kit/assets"
```

### 4. Lookup and derivation helpers

```go
chars.Len()                            // number of characters
chars.All()                            // all characters in definition order (a copy)
chars.GetCharacter("ZUNDAMON")         // case-insensitive lookup by ID; nil if absent
chars.GetDefault()                     // the is_default character; nil if none
chars.GetCharacterWithDefault("xxx")   // falls back to the default when not found
char.ReferenceURLFor("9:16")           // reference URL for the aspect ratio; reference_url if none

// Derive a new collection with only zundamon's seed replaced (the original is unchanged)
overridden := chars.WithSeedOverride("zundamon", 999)
```

---

## 📦 Package Structure

| Package | Contents |
| --- | --- |
| `github.com/shouni/go-character-kit/character` | The character domain model, construction, JSON parsing, validation and lookup helpers. |
| `github.com/shouni/go-character-kit/assets` | Loading of the bundled, `go:embed`-ed character definition JSON. |

---

## 🛡 Validation

`NewCharacters` / `ParseCharacters` catch the following configuration mistakes at construction:

* An empty character definition (empty array or nil)
* Missing `id` / `name` / `visual_cues`
* Leading or trailing whitespace in `reference_url` (a whitespace-only value included; leaving it out is fine)
* Leading or trailing whitespace in `id`
* Duplicate IDs, compared case-insensitively
* Malformed aspect-ratio keys in `reference_urls` (only the `"16:9"` form is allowed, which catches typos such as `"16x9"`)
* Empty URLs in `reference_urls`
* Unknown keys in `sheets` (only `expressions` / `outfit` / `parts`), and empty or padded URLs in it
* More than one `is_default` (the offending IDs are listed in the error message)

---

## 📜 License

This project is released under the [MIT License](LICENSE).
