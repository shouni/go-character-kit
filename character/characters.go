package character

import (
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

// Characters は定義順を保持するキャラクター集合です。
// NewCharacters / ParseCharacters で構築した時点で検証済みであることが保証され、
// 以降は読み取り専用として扱えます。各メソッドは nil レシーバーでも安全に動作します。
type Characters struct {
	list        []Character
	byID        map[string]*Character // 小文字化したIDをキーにした検索用マップ
	defaultChar *Character
}

// NewCharacters はキャラクター定義リストを検証し、集合を構築して返します。
// list は複製して保持するため、呼び出し後に list を変更しても集合には影響しません。
func NewCharacters(list []Character) (*Characters, error) {
	if err := validateList(list); err != nil {
		return nil, err
	}
	return newValidated(list), nil
}

// ParseCharacters はJSONバイト列からキャラクター定義をパースして返します。
func ParseCharacters(charactersJSON []byte) (*Characters, error) {
	var list []Character
	if err := json.Unmarshal(charactersJSON, &list); err != nil {
		return nil, fmt.Errorf("キャラクター情報のJSONパースに失敗しました: %w", err)
	}

	return NewCharacters(list)
}

// newValidated は検証済みのリストから集合を構築します。
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

// Len は登録されているキャラクター数を返します。
func (c *Characters) Len() int {
	if c == nil {
		return 0
	}
	return len(c.list)
}

// All は定義順のキャラクター一覧をコピーで返します。
// 返り値を変更しても集合には影響しません。
func (c *Characters) All() []Character {
	if c == nil {
		return nil
	}
	return cloneList(c.list)
}

// GetCharacter は、指定されたIDからキャラクター情報を特定します。ID照合は大小文字を
// 無視します。見つかった場合はそのコピーへのポインタを、存在しない場合は nil を返します。
func (c *Characters) GetCharacter(id string) *Character {
	if c == nil {
		return nil
	}
	char, ok := c.byID[strings.ToLower(id)]
	if !ok {
		return nil
	}
	cloned := char.clone()
	return &cloned
}

// GetDefault は IsDefault が true のキャラクターのコピーを返します。
// 存在しない場合は nil を返します。
func (c *Characters) GetDefault() *Character {
	if c == nil || c.defaultChar == nil {
		return nil
	}
	cloned := c.defaultChar.clone()
	return &cloned
}

// GetCharacterWithDefault は、指定されたIDでキャラクターを検索し、見つからない場合は
// デフォルトのキャラクターを返します。どちらも見つからない場合は nil を返します。
func (c *Characters) GetCharacterWithDefault(id string) *Character {
	if char := c.GetCharacter(id); char != nil {
		return char
	}
	return c.GetDefault()
}

// WithSeedOverride は、id に一致するキャラクターの Seed だけを seed に差し替えた
// 新しい集合を返します。レシーバー自体は変更されません。ID照合は GetCharacter と
// 同じく大小文字を無視し、一致するキャラクターがいない場合はレシーバーをそのまま返します。
func (c *Characters) WithSeedOverride(id string, seed int64) *Characters {
	if c == nil {
		return nil
	}
	if _, ok := c.byID[strings.ToLower(id)]; !ok {
		return c
	}
	list := c.All()
	for i := range list {
		if strings.EqualFold(list[i].ID, id) {
			overridden := seed
			list[i].Seed = &overridden
		}
	}
	return newValidated(list)
}

// aspectRatioKeyPattern は reference_urls のキーとして許可する "16:9" 形式です。
var aspectRatioKeyPattern = regexp.MustCompile(`^\d+:\d+$`)

// validateList はキャラクター定義リストの設定ミスを検出します。
func validateList(list []Character) error {
	if len(list) == 0 {
		return fmt.Errorf("キャラクター定義が空です")
	}
	seen := make(map[string]struct{}, len(list))
	var defaultIDs []string
	for i, char := range list {
		id := strings.TrimSpace(char.ID)
		if id == "" {
			return fmt.Errorf("キャラクターIDが空です (index: %d)", i)
		}
		if char.ID != id {
			return fmt.Errorf("キャラクターIDに前後の空白があります: %q", char.ID)
		}
		key := strings.ToLower(id)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("キャラクターIDが重複しています: %s", id)
		}
		seen[key] = struct{}{}
		if strings.TrimSpace(char.Name) == "" {
			return fmt.Errorf("キャラクター名が空です (id: %s)", id)
		}
		if strings.TrimSpace(char.ReferenceURL) == "" {
			return fmt.Errorf("参照画像URLが空です (id: %s)", id)
		}
		if len(char.VisualCues) == 0 {
			return fmt.Errorf("visual_cuesが空です (id: %s)", id)
		}
		for _, ratio := range slices.Sorted(maps.Keys(char.ReferenceURLs)) {
			if !aspectRatioKeyPattern.MatchString(ratio) {
				return fmt.Errorf("reference_urlsのアスペクト比キーが不正です (id: %s, key: %q): \"16:9\" のような形式で指定してください", id, ratio)
			}
			if strings.TrimSpace(char.ReferenceURLs[ratio]) == "" {
				return fmt.Errorf("reference_urlsのURLが空です (id: %s, key: %s)", id, ratio)
			}
		}
		if char.IsDefault {
			defaultIDs = append(defaultIDs, id)
		}
	}
	if len(defaultIDs) > 1 {
		return fmt.Errorf("デフォルトキャラクターが複数あります: %s", strings.Join(defaultIDs, ", "))
	}
	return nil
}
