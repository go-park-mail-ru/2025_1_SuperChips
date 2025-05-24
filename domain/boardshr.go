package domain

import "time"

// Приглашения могут иметь следующие параметры:
// - Names:     Имена приглашаемых пользователей. При отсутствии параметра ссылка публичная.
// - TimeLimit: Время, в течение которого ссылка активна. При отсутствии параметра ссылка бессрочная.
// - UsageLimit: Количество пользователей, которые могут воспользоваться этой ссылкой. При отсутствии параметра количество не ограничено. При неоднократном становлении соавтором одним и тем же пользователем лимит не растрачивается.
//
//easyjson:json
type Invitaion struct {
	Names      *[]string `json:"names,omitempty"`
	TimeLimit  *time.Time  `json:"time_limit,omitempty"`
	UsageLimit *int      `json:"usage_limit,omitempty"`
}

//easyjson:json
type LinkParams struct {
	Link       string     `json:"link"`
	Names      *[]string  `json:"names"`
	TimeLimit  *time.Time `json:"time_limit"`
	UsageLimit *int64     `json:"usage_limit"`
	UsageCount int64     `json:"usage_count"`
}

//easyjson:json
type BodyWithUsername struct {
	Name string `json:"name"`
}
