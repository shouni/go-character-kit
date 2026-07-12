package character

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCharactersBuildsListAndLookup(t *testing.T) {
	t.Parallel()

	seed := int64(10001)
	chars, err := NewCharacters([]Character{
		{
			ID:           "zundamon",
			Name:         "Zundamon",
			ReferenceURL: "gs://bucket/zundamon.png",
			VisualCues:   []string{"green hair"},
			Seed:         &seed,
			IsDefault:    true,
		},
	})

	require.NoError(t, err)
	assert.Len(t, chars.List, 1)
	assert.Equal(t, "zundamon", chars.GetCharacter("ZUNDAMON").ID)
	assert.Equal(t, "zundamon", chars.GetDefault().ID)
	if assert.NotNil(t, chars.GetDefault().Seed) {
		assert.Equal(t, int64(10001), *chars.GetDefault().Seed)
	}
	assert.Equal(t, "zundamon", chars.GetCharacterWithDefault("").ID)
}

func TestParseCharactersBuildsListAndLookup(t *testing.T) {
	t.Parallel()

	chars, err := ParseCharacters([]byte(`[
		{
			"id": "zundamon",
			"name": "Zundamon",
			"reference_url": "gs://bucket/zundamon.png",
			"visual_cues": ["green hair"],
			"seed": 10001,
			"is_default": true
		}
	]`))

	require.NoError(t, err)
	assert.Len(t, chars.List, 1)
	assert.Equal(t, "zundamon", chars.GetCharacter("ZUNDAMON").ID)
	assert.Equal(t, "zundamon", chars.GetDefault().ID)
	if assert.NotNil(t, chars.GetDefault().Seed) {
		assert.Equal(t, int64(10001), *chars.GetDefault().Seed)
	}
	assert.Equal(t, "zundamon", chars.GetCharacterWithDefault("").ID)
}

func TestCharacterReferenceURLForPrefersAspectRatioMatch(t *testing.T) {
	t.Parallel()

	char := Character{
		ID:           "tsumugi",
		ReferenceURL: "gs://bucket/tsumugi-16x9.png",
		ReferenceURLs: map[string]string{
			"9:16": "gs://bucket/tsumugi-9x16.png",
			"1:1":  "gs://bucket/tsumugi-1x1.png",
		},
	}

	assert.Equal(t, "gs://bucket/tsumugi-9x16.png", char.ReferenceURLFor("9:16"))
	assert.Equal(t, "gs://bucket/tsumugi-1x1.png", char.ReferenceURLFor("1:1"))
}

func TestCharacterReferenceURLForFallsBackWhenAspectRatioMissing(t *testing.T) {
	t.Parallel()

	char := Character{
		ID:           "tsumugi",
		ReferenceURL: "gs://bucket/tsumugi-16x9.png",
		ReferenceURLs: map[string]string{
			"9:16": "gs://bucket/tsumugi-9x16.png",
		},
	}

	// No "4:3" entry, and no aspect ratio requested at all: both fall back to ReferenceURL.
	assert.Equal(t, "gs://bucket/tsumugi-16x9.png", char.ReferenceURLFor("4:3"))
	assert.Equal(t, "gs://bucket/tsumugi-16x9.png", char.ReferenceURLFor(""))
}

func TestCharacterReferenceURLForNilCharacterReturnsEmpty(t *testing.T) {
	t.Parallel()

	var char *Character
	assert.Equal(t, "", char.ReferenceURLFor("9:16"))
}

func TestParseCharactersRejectsDuplicateIDs(t *testing.T) {
	t.Parallel()

	_, err := ParseCharacters([]byte(`[
		{
			"id": "zundamon",
			"name": "Zundamon",
			"reference_url": "gs://bucket/zundamon.png",
			"visual_cues": ["green hair"]
		},
		{
			"id": "ZUNDAMON",
			"name": "Zundamon 2",
			"reference_url": "gs://bucket/zundamon-2.png",
			"visual_cues": ["green hair"]
		}
	]`))

	assert.Error(t, err)
}
