// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package rest

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

func easyjson9b8f5552DecodeGithubComGoParkMailRu20251SuperChipsInternalRest(in *jlexer.Lexer, out *Username) {
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
		case "username":
			out.Username = string(in.String())
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
func easyjson9b8f5552EncodeGithubComGoParkMailRu20251SuperChipsInternalRest(out *jwriter.Writer, in Username) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"username\":"
		out.RawString(prefix[1:])
		out.String(string(in.Username))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Username) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9b8f5552EncodeGithubComGoParkMailRu20251SuperChipsInternalRest(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Username) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9b8f5552EncodeGithubComGoParkMailRu20251SuperChipsInternalRest(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Username) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9b8f5552DecodeGithubComGoParkMailRu20251SuperChipsInternalRest(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Username) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9b8f5552DecodeGithubComGoParkMailRu20251SuperChipsInternalRest(l, v)
}
