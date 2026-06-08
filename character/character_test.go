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
