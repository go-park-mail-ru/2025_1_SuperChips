package domain


// это несоответствие имен структуры и json-тегов внутри нее
// временное решение, чтобы не ломать фронтенд
// как только фронтенд будет готов к изменениям, можно будет
// поменять имена полей и теги
type PinData struct {
	FlowID      uint64 `json:"-" db:"flow_id"`
	Header      string `json:"header,omitempty" db:"title"`
	Description string `json:"description,omitempty" db:"description"`
	MediaURL    string `json:"image,omitempty" db:"media_url"`
	AuthorID    uint64 `json:"author_id" db:"author_id"`
	IsPrivate   bool   `json:"-" db:"is_private"`
	Created_at  string `json:"-" db:"create_at"`
	Updated_at  string `json:"-" db:"updated_at"`
}
