package header

type Metadata struct {
	DeviceID       string              `json:"device_id,omitempty"`
	DeviceOS       string              `json:"device_os,omitempty"`
	Platform       string              `json:"platform,omitempty"`
	UserAgent      string              `json:"user_agent,omitempty"`
	AppVersion     string              `json:"app_version,omitempty"`
	Authorization  string              `json:"authorization,omitempty"`
	Density        string              `json:"density,omitempty"`
	Cookies        map[string][]string `json:"cookies,omitempty"`
	Extras         map[string][]string `json:"extras,omitempty"`
	ZLPToken       string              `json:"zlp_token,omitempty"`
	AccessToken    string              `json:"access_token,omitempty"`
	UserID         string              `json:"user_id,omitempty"`
	AcceptLanguage string              `json:"language,omitempty"`
	UserIP         string              `json:"user_ip,omitempty"`
}
