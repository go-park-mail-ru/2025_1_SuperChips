package domain

import "html"

type Answer struct {
	Type       string `json:"type"`
	Content    string `json:"content"`
	QuestionID int    `json:"question_id"`
}

type Question struct {
	ID    uint64 `json:"question_id"`
	Text  string `json:"text"`
	Order int64  `json:"order"`
	Type  string `json:"type"`
}

type Poll struct {
	ID        uint64     `json:"id"`
	Header    string     `json:"header"`
	Questions []Question `json:"questions,omitempty"`
	Delay     int        `json:"delay"`
	Screen    []string   `json:"screen"`
}

func (q *Question) Escape() {
	q.Text = html.EscapeString(q.Text)
}

func (p *Poll) Escape() {
	p.Header = html.EscapeString(p.Header)
}
