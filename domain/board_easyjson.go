// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package domain

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson202377feDecodeGithubComGoParkMailRu20251SuperChipsDomain(in *jlexer.Lexer, out *UpdateData) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "name":
			out.Name = string(in.String())
		case "is_private":
			out.IsPrivate = bool(in.Bool())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson202377feEncodeGithubComGoParkMailRu20251SuperChipsDomain(out *jwriter.Writer, in UpdateData) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix[1:])
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"is_private\":"
		out.RawString(prefix)
		out.Bool(bool(in.IsPrivate))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v UpdateData) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson202377feEncodeGithubComGoParkMailRu20251SuperChipsDomain(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v UpdateData) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson202377feEncodeGithubComGoParkMailRu20251SuperChipsDomain(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *UpdateData) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson202377feDecodeGithubComGoParkMailRu20251SuperChipsDomain(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *UpdateData) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson202377feDecodeGithubComGoParkMailRu20251SuperChipsDomain(l, v)
}
func easyjson202377feDecodeGithubComGoParkMailRu20251SuperChipsDomain1(in *jlexer.Lexer, out *BoardRequest) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "flow_id":
			out.FlowID = int(in.Int())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson202377feEncodeGithubComGoParkMailRu20251SuperChipsDomain1(out *jwriter.Writer, in BoardRequest) {
	out.RawByte('{')
	first := true
	_ = first
	if in.FlowID != 0 {
		const prefix string = ",\"flow_id\":"
		first = false
		out.RawString(prefix[1:])
		out.Int(int(in.FlowID))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v BoardRequest) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson202377feEncodeGithubComGoParkMailRu20251SuperChipsDomain1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v BoardRequest) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson202377feEncodeGithubComGoParkMailRu20251SuperChipsDomain1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *BoardRequest) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson202377feDecodeGithubComGoParkMailRu20251SuperChipsDomain1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *BoardRequest) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson202377feDecodeGithubComGoParkMailRu20251SuperChipsDomain1(l, v)
}
func easyjson202377feDecodeGithubComGoParkMailRu20251SuperChipsDomain2(in *jlexer.Lexer, out *Board) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.ID = int(in.Int())
		case "author_id":
			out.AuthorID = int(in.Int())
		case "author_username":
			out.AuthorUsername = string(in.String())
		case "name":
			out.Name = string(in.String())
		case "is_editable":
			out.IsEditable = bool(in.Bool())
		case "is_private":
			out.IsPrivate = bool(in.Bool())
		case "flow_count":
			out.FlowCount = int(in.Int())
		case "preview":
			if in.IsNull() {
				in.Skip()
				out.Preview = nil
			} else {
				in.Delim('[')
				if out.Preview == nil {
					if !in.IsDelim(']') {
						out.Preview = make([]PinData, 0, 0)
					} else {
						out.Preview = []PinData{}
					}
				} else {
					out.Preview = (out.Preview)[:0]
				}
				for !in.IsDelim(']') {
					var v1 PinData
					(v1).UnmarshalEasyJSON(in)
					out.Preview = append(out.Preview, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "gradient":
			if in.IsNull() {
				in.Skip()
				out.Gradient = nil
			} else {
				in.Delim('[')
				if out.Gradient == nil {
					if !in.IsDelim(']') {
						out.Gradient = make([]string, 0, 4)
					} else {
						out.Gradient = []string{}
					}
				} else {
					out.Gradient = (out.Gradient)[:0]
				}
				for !in.IsDelim(']') {
					var v2 string
					v2 = string(in.String())
					out.Gradient = append(out.Gradient, v2)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson202377feEncodeGithubComGoParkMailRu20251SuperChipsDomain2(out *jwriter.Writer, in Board) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.Int(int(in.ID))
	}
	{
		const prefix string = ",\"author_id\":"
		out.RawString(prefix)
		out.Int(int(in.AuthorID))
	}
	if in.AuthorUsername != "" {
		const prefix string = ",\"author_username\":"
		out.RawString(prefix)
		out.String(string(in.AuthorUsername))
	}
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"is_editable\":"
		out.RawString(prefix)
		out.Bool(bool(in.IsEditable))
	}
	{
		const prefix string = ",\"is_private\":"
		out.RawString(prefix)
		out.Bool(bool(in.IsPrivate))
	}
	{
		const prefix string = ",\"flow_count\":"
		out.RawString(prefix)
		out.Int(int(in.FlowCount))
	}
	if len(in.Preview) != 0 {
		const prefix string = ",\"preview\":"
		out.RawString(prefix)
		{
			out.RawByte('[')
			for v3, v4 := range in.Preview {
				if v3 > 0 {
					out.RawByte(',')
				}
				(v4).MarshalEasyJSON(out)
			}
			out.RawByte(']')
		}
	}
	if len(in.Gradient) != 0 {
		const prefix string = ",\"gradient\":"
		out.RawString(prefix)
		{
			out.RawByte('[')
			for v5, v6 := range in.Gradient {
				if v5 > 0 {
					out.RawByte(',')
				}
				out.String(string(v6))
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Board) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson202377feEncodeGithubComGoParkMailRu20251SuperChipsDomain2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Board) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson202377feEncodeGithubComGoParkMailRu20251SuperChipsDomain2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Board) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson202377feDecodeGithubComGoParkMailRu20251SuperChipsDomain2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Board) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson202377feDecodeGithubComGoParkMailRu20251SuperChipsDomain2(l, v)
}
