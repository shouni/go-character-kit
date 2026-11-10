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
	ReferenceURL string `json:"reference_url"`
	// ReferenceURLs holds reference image URLs keyed by an aspect-ratio string such as
	// "1:1", "9:16" or "16:9". When the reference image and the target output differ greatly in
	// aspect ratio (for example, a wide three-pose sheet used as the reference for a portrait
	// keyframe), details such as colours, prop placement and hair tend to drift from one
	// generation to the next, so callers are expected to prefer the entry matching the aspect
	// ratio being generated. When there is no matching entry, or the field is unset, it falls
	// back to ReferenceURL.
	ReferenceURLs map[string]string `json:"reference_urls,omitempty"`
	Seed          *int64            `json:"seed,omitempty"` // optional generation seed tied to the character
	IsDefault     bool              `json:"is_default"`     // whether this character's seed takes priority as the page-wide seed
}

// ReferenceURLFor returns the reference image URL matching aspectRatio if there is one, and
// ReferenceURL (the default fallback) otherwise. It also returns ReferenceURL when aspectRatio
// is empty.
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
