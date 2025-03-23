package domain


// это несоответствие имен структуры и json-тегов внутри нее
// временное решение, чтобы не ломать фронтенд
// как только фронтенд будет готов к изменениям, можно будет
// поменять имена полей и теги
type PinData struct {
	FlowID      uint64 `json:"-"`
	Header      string `json:"header,omitempty"`
	Description string `json:"description,omitempty"`
	MediaURL    string `json:"image,omitempty"`
	AuthorID    uint64 `json:"author_id"`
	IsPrivate   bool   `json:"-"`
	Created_at  string `json:"-"`
	Updated_at  string `json:"-"`
}
