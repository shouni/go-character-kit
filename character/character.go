// Package character provides the type for character definitions injected into
// generation prompts, along with their lookup and validation.
package character

import (
	"maps"
	"slices"
)

// Character holds the definition of a character appearing in a comic.
type Character struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	VisualCues []string `json:"visual_cues"` // visual traits injected into the generation prompt
	// ReferenceURL is the reference image URL used for consistency (the aspect-agnostic default
	// fallback). It is used when ReferenceURLs has no entry for the aspect ratio the caller asks for.
	//
	// It may be empty: a new character has no image yet, and her first design sheet has to be
	// drawn from VisualCues alone. A placeholder image would be worse than none, because generation
	// prompts treat the attached reference as the ground truth for the character's look.
	ReferenceURL string `json:"reference_url,omitempty"`
	// ReferenceURLs holds reference image URLs keyed by an aspect-ratio string such as
	// "1:1", "9:16" or "16:9". When the reference image and the target output differ greatly in
	// aspect ratio (for example, a wide three-pose sheet used as the reference for a portrait
	// keyframe), details such as colours, prop placement and hair tend to drift from one
	// generation to the next, so callers are expected to prefer the entry matching the aspect
	// ratio being generated. When there is no matching entry, or the field is unset, it falls
	// back to ReferenceURL.
	ReferenceURLs map[string]string `json:"reference_urls,omitempty"`
	// Sheets holds supplementary design sheets keyed by kind (SheetExpressions, SheetOutfit,
	// SheetParts). They sit beside the reference image rather than replacing it: the reference is the
	// full-body identity anchor, and a sheet shows one aspect of the character on its own — the face
	// in several moods, the costume's construction, accessories laid out singly. Keeping them apart
	// is what lets the anchor stay full-body only, which design-sheet generation requires (insets and
	// detached parts on an anchor made the model split or fuse figures). A caller attaches the sheets
	// it needs next to the reference, e.g. the parts sheet when an accessory keeps drifting.
	Sheets    map[string]string `json:"sheets,omitempty"`
	Seed      *int64            `json:"seed,omitempty"` // optional generation seed tied to the character
	IsDefault bool              `json:"is_default"`     // whether this character's seed takes priority as the page-wide seed
}

// The kinds of supplementary design sheet a character can have (Character.Sheets keys). A key
// outside this list is rejected, as a misspelt field name is: "part" or "expression" would
// otherwise load cleanly and the sheet would just never be found.
const (
	SheetExpressions = "expressions" // the same face in several moods, head and shoulders
	SheetOutfit      = "outfit"      // the costume shown clearly, front and back
	SheetParts       = "parts"       // accessories and props laid out singly
)

// sheetKinds lists every supplementary sheet kind, in the order they are usually generated.
var sheetKinds = []string{SheetExpressions, SheetOutfit, SheetParts}

// SheetKinds returns every supplementary sheet kind (a copy, so a caller cannot change what
// validation accepts).
func SheetKinds() []string {
	return slices.Clone(sheetKinds)
}

// SheetURL returns the URL of the supplementary design sheet of the given kind, or "" when the
// character has none.
func (c *Character) SheetURL(kind string) string {
	if c == nil {
		return ""
	}
	return c.Sheets[kind]
}

// ReferenceURLFor returns the reference image URL matching aspectRatio if there is one, and
// ReferenceURL (the default fallback) otherwise. It also returns ReferenceURL when aspectRatio
// is empty. The result is empty for a character with no reference image yet, so callers must
// treat "" as "draw from the visual cues alone", not as an image to attach.
func (c *Character) ReferenceURLFor(aspectRatio string) string {
	if c == nil {
		return ""
	}
	if aspectRatio != "" {
		if url, ok := c.ReferenceURLs[aspectRatio]; ok && url != "" {
			return url
		}
	}
	return c.ReferenceURL
}

// clone returns a copy that also duplicates the reference-typed fields.
func (c Character) clone() Character {
	c.VisualCues = slices.Clone(c.VisualCues)
	c.ReferenceURLs = maps.Clone(c.ReferenceURLs)
	c.Sheets = maps.Clone(c.Sheets)
	if c.Seed != nil {
		seed := *c.Seed
		c.Seed = &seed
	}
	return c
}

// cloneList returns a copy of list with every element deep-copied.
func cloneList(list []Character) []Character {
	if list == nil {
		return nil
	}
	cloned := make([]Character, len(list))
	for i := range list {
		cloned[i] = list[i].clone()
	}
	return cloned
}
