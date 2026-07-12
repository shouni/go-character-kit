package character

// Character は漫画に登場するキャラクターの定義を保持します。
type Character struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	VisualCues []string `json:"visual_cues"` // 生成プロンプトに注入する外見上の特徴
	// ReferenceURL は一貫性保持のための参照画像URL（アスペクト比を問わない既定のフォールバック）。
	// ReferenceURLs に呼び出し側の要求するアスペクト比のエントリがない場合に使われます。
	ReferenceURL string `json:"reference_url"`
	// ReferenceURLs は "1:1"/"9:16"/"16:9" などアスペクト比文字列をキーにした参照画像URLです。
	// 参照画像とターゲットの生成物のアスペクト比が大きく異なる（例: 横長3ポーズシートを縦長
	// キーフレームの参照に使う）と、色・小物配置・髪型などの細部が生成のたびにブレやすいため、
	// 呼び出し側が生成対象と同じアスペクト比のエントリを優先的に使うことを想定しています。
	// 該当エントリが無い、またはこのフィールド自体が未設定の場合は ReferenceURL にフォールバックします。
	ReferenceURLs map[string]string `json:"reference_urls,omitempty"`
	Seed          *int64            `json:"seed,omitempty"` // キャラクターに紐づく任意の生成シード
	IsDefault     bool              `json:"is_default"`     // ページ全体の代表Seedとして優先するか
}

// ReferenceURLFor は、aspectRatio に一致する参照画像URLがあればそれを返し、無ければ
// ReferenceURL（既定のフォールバック）を返します。aspectRatio が空文字の場合も同様に
// ReferenceURL を返します。
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

// Characters は表示順を持つリストとID検索用マップを保持します。
type Characters struct {
	List []Character
	ByID map[string]*Character
}
