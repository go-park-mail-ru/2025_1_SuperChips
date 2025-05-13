package domain

//easyjson:json
type PinDataUpdate struct {
	FlowID      *uint64 `json:"flow_id,omitempty"`
	Header      *string `json:"header,omitempty"`
	Description *string `json:"description,omitempty"`
	IsPrivate   *bool   `json:"is_private,omitempty"`
}

type PinDataCreate struct {
	Header      string `json:"header,omitempty"`
	Description string `json:"description,omitempty"`
	IsPrivate   bool   `json:"is_private,omitempty"`
	Colors      []string
	Width       int
	Height      int
}
