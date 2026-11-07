package mwclientmini

type RecentChange struct {
	Type       string `json:"type"`
	Title      string `json:"title"`
	User       string `json:"user"`
	Bot        bool   `json:"bot"`
	Minor      bool   `json:"minor"`
	ServerName string `json:"server_name"`
	Wiki       string `json:"wiki"`
	Namespace  int    `json:"namespace"`
	Timestamp  int64  `json:"timestamp"`
	Comment    string `json:"comment"`

	LogType   string `json:"log_type"`
	LogAction string `json:"log_action"`

	Meta struct {
		Domain string `json:"domain"`
	} `json:"meta"`
}
