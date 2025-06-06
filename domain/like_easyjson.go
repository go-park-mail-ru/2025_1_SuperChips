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

func easyjson52421b6dDecodeGithubComGoParkMailRu20251SuperChipsDomain(in *jlexer.Lexer, out *Like) {
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
		case "pin_id":
			out.PinID = int(in.Int())
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
func easyjson52421b6dEncodeGithubComGoParkMailRu20251SuperChipsDomain(out *jwriter.Writer, in Like) {
	out.RawByte('{')
	first := true
	_ = first
	if in.PinID != 0 {
		const prefix string = ",\"pin_id\":"
		first = false
		out.RawString(prefix[1:])
		out.Int(int(in.PinID))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Like) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson52421b6dEncodeGithubComGoParkMailRu20251SuperChipsDomain(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Like) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson52421b6dEncodeGithubComGoParkMailRu20251SuperChipsDomain(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Like) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson52421b6dDecodeGithubComGoParkMailRu20251SuperChipsDomain(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Like) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson52421b6dDecodeGithubComGoParkMailRu20251SuperChipsDomain(l, v)
}
