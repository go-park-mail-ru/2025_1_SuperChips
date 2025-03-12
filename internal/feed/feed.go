package feed

type PinData struct {
	Header string `json:"header"` // Заголовок пина
	Image  string `json:"image"`  // URL пина на FileServer'e бэка
	Author string `json:"author"` // Автор пина
}
