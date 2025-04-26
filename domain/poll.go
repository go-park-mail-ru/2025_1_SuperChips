package domain

import "html"

type Answer struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type Question struct {
	ID    int    `json:"question_id"`
	Text  string `json:"text"`
	Order int    `json:"order"`
	Type  string `json:"type"`
}

type Poll struct {
	ID        int        `json:"id"`
	Header    string     `json:"header"`
	Questions []Question `json:"questions,omitempty"`
}

func (q *Question) Escape() {
	q.Text = html.EscapeString(q.Text)
}

func (p *Poll) Escape() {
	p.Header = html.EscapeString(p.Header)
}
