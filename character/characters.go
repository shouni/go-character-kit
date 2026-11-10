package character

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"regexp"
	"slices"
	"strings"
)

// Characters is a collection of characters that keeps their definition order.
// A collection built by NewCharacters or ParseCharacters is guaranteed to be validated and
// can be treated as read-only from then on. Every method is safe to call on a nil receiver.
type Characters struct {
	list        []Character
	byID        map[string]*Character // lookup map keyed by the lower-cased ID
	defaultChar *Character
}

// NewCharacters validates a list of character definitions and returns the collection built
// from it. The list is copied, so changing it after the call does not affect the collection.
func NewCharacters(list []Character) (*Characters, error) {
	if err := validateList(list); err != nil {
		return nil, err
	}
	return newValidated(list), nil
}

// ParseCharacters parses character definitions from JSON bytes and returns the collection.
//
// A field name not in the definition is an error. Silently skipping a typo such as "sead" or
// "referense_urls" only leaves Seed nil or drops the ratio-specific reference to the default,
// and the generated output changes without a word. Field names are checked for the same reason
// the shape of the reference_urls keys is.
func ParseCharacters(charactersJSON []byte) (*Characters, error) {
	dec := json.NewDecoder(bytes.NewReader(charactersJSON))
	dec.DisallowUnknownFields()

	var list []Character
	if err := dec.Decode(&list); err != nil {
		return nil, fmt.Errorf("failed to parse character JSON: %w", err)
	}
	// Decode reads only the first value, so make sure nothing follows it.
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("character JSON has trailing content after the array")
	}

	return NewCharacters(list)
}

// newValidated builds a collection from an already validated list.
func newValidated(list []Character) *Characters {
	c := &Characters{
		list: cloneList(list),
		byID: make(map[string]*Character, len(list)),
	}
	for i := range c.list {
		char := &c.list[i]
		c.byID[strings.ToLower(char.ID)] = char
		if char.IsDefault {
			c.defaultChar = char
		}
	}
	return c
}

// Len returns the number of characters in the collection.
func (c *Characters) Len() int {
	if c == nil {
		return 0
	}
	return len(c.list)
}

// All returns a copy of the characters in definition order.
// Changing the result does not affect the collection.
func (c *Characters) All() []Character {
	if c == nil {
		return nil
	}
	return cloneList(c.list)
}

// GetCharacter looks up the character with the given ID. Matching ignores case and
// surrounding whitespace. It returns a pointer to a copy of the character when found, and nil
// otherwise.
//
// Surrounding whitespace is ignored because IDs come from AI output such as recipes and
// scripts, and from human input. Definition-side IDs may not carry surrounding whitespace, so
// trimming the lookup key cannot match the wrong character. Without it, every caller would
// write GetCharacter(strings.TrimSpace(id)).
func (c *Characters) GetCharacter(id string) *Character {
	if c == nil {
		return nil
	}
	char, ok := c.byID[lookupKey(id)]
	if !ok {
		return nil
	}
	cloned := char.clone()
	return &cloned
}

// GetDefault returns a copy of the character whose IsDefault is true.
// It returns nil when there is none.
func (c *Characters) GetDefault() *Character {
	if c == nil || c.defaultChar == nil {
		return nil
	}
	cloned := c.defaultChar.clone()
	return &cloned
}

// GetCharacterWithDefault looks up the character with the given ID and returns the default
// character when it is not found. It returns nil when neither exists.
func (c *Characters) GetCharacterWithDefault(id string) *Character {
	if char := c.GetCharacter(id); char != nil {
		return char
	}
	return c.GetDefault()
}

// WithSeedOverride returns a new collection in which only the Seed of the character matching
// id is replaced with seed. The receiver itself is not modified. Like GetCharacter, matching
// ignores case; when no character matches, the receiver is returned as is.
func (c *Characters) WithSeedOverride(id string, seed int64) *Characters {
	if c == nil {
		return nil
	}
	if _, ok := c.byID[lookupKey(id)]; !ok {
		return c
	}
	list := c.All()
	for i := range list {
		if lookupKey(list[i].ID) == lookupKey(id) {
			overridden := seed
			list[i].Seed = &overridden
		}
	}
	return newValidated(list)
}

// lookupKey returns the key used to match IDs. Both the definition side and the lookup side go through it.
func lookupKey(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

// aspectRatioKeyPattern is the "16:9" form allowed as a reference_urls key.
var aspectRatioKeyPattern = regexp.MustCompile(`^\d+:\d+$`)

// validateList detects configuration mistakes in a list of character definitions.
func validateList(list []Character) error {
	if len(list) == 0 {
		return fmt.Errorf("character definitions are empty")
	}
	seen := make(map[string]struct{}, len(list))
	var defaultIDs []string
	for i, char := range list {
		id := strings.TrimSpace(char.ID)
		if id == "" {
			return fmt.Errorf("character ID is empty (index: %d)", i)
		}
		if char.ID != id {
			return fmt.Errorf("character ID has leading or trailing whitespace: %q", char.ID)
		}
		key := strings.ToLower(id)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate character ID: %s", id)
		}
		seen[key] = struct{}{}
		if strings.TrimSpace(char.Name) == "" {
			return fmt.Errorf("character name is empty (id: %s)", id)
		}
		if err := validateURL(char.ReferenceURL, "reference_url", fmt.Sprintf("id: %s", id)); err != nil {
			return err
		}
		if len(char.VisualCues) == 0 {
			return fmt.Errorf("visual_cues is empty (id: %s)", id)
		}
		for _, ratio := range slices.Sorted(maps.Keys(char.ReferenceURLs)) {
			if !aspectRatioKeyPattern.MatchString(ratio) {
				return fmt.Errorf("invalid reference_urls aspect-ratio key (id: %s, key: %q): use the form \"16:9\"", id, ratio)
			}
			if err := validateURL(char.ReferenceURLs[ratio], "reference_urls URL", fmt.Sprintf("id: %s, key: %s", id, ratio)); err != nil {
				return err
			}
		}
		if char.IsDefault {
			defaultIDs = append(defaultIDs, id)
		}
	}
	if len(defaultIDs) > 1 {
		return fmt.Errorf("more than one default character: %s", strings.Join(defaultIDs, ", "))
	}
	return nil
}

// validateURL checks that a reference image URL is non-empty and has no leading or trailing
// whitespace.
//
// Whitespace is rejected rather than trimmed so that URLs are treated the same as IDs. Rejecting
// it here guarantees that ReferenceURLFor on a validated Character never returns an empty or
// padded string, so callers need not TrimSpace the result before checking it.
func validateURL(url, what, where string) error {
	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		return fmt.Errorf("%s is empty (%s)", what, where)
	}
	if url != trimmed {
		return fmt.Errorf("%s has leading or trailing whitespace (%s): %q", what, where, url)
	}
	return nil
}
