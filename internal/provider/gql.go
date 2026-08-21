package provider

const monitorSelection = `
	id
	name
	description
	tags
	type
	enabled
	interval_seconds
	timeout_seconds
	ip_family
	target
	method
	requestHeaders { key value }
	request_body
	dns_query_name
	dns_query_type
	heartbeat_token
	heartbeatUrl
	heartbeatStartUrl
	heartbeatFinishUrl
	heartbeatErrorUrl
	statusBadgeUrl
	statusBadgeJsonUrl
	uptimeBadgeUrl
	uptimeBadgeJsonUrl
	latencyBadgeUrl
	latencyBadgeJsonUrl
	badgeMarkdown
	follow_redirects
	verify_tls
	proxy_url
	status
	retention_days
	uptime {
		oneHour
		twentyFourHours
		sevenDays
		thirtyDays
	}
	conditions { expression }
	probes { id }
	notificationChannels { id }
`

type gqlKeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type gqlUptime struct {
	OneHour         *float64 `json:"oneHour"`
	TwentyFourHours *float64 `json:"twentyFourHours"`
	SevenDays       *float64 `json:"sevenDays"`
	ThirtyDays      *float64 `json:"thirtyDays"`
}

type gqlMonitor struct {
	ID                   string         `json:"id"`
	Name                 string         `json:"name"`
	Description          *string        `json:"description"`
	Tags                 []string       `json:"tags"`
	Type                 string         `json:"type"`
	Enabled              bool           `json:"enabled"`
	IntervalSeconds      int64          `json:"interval_seconds"`
	TimeoutSeconds       int64          `json:"timeout_seconds"`
	IPFamily             string         `json:"ip_family"`
	Target               string         `json:"target"`
	Method               *string        `json:"method"`
	RequestHeaders       []gqlKeyValue  `json:"requestHeaders"`
	RequestBody          *string        `json:"request_body"`
	DNSQueryName         *string        `json:"dns_query_name"`
	DNSQueryType         *string        `json:"dns_query_type"`
	HeartbeatToken       *string        `json:"heartbeat_token"`
	HeartbeatURL         *string        `json:"heartbeatUrl"`
	HeartbeatStartURL    *string        `json:"heartbeatStartUrl"`
	HeartbeatFinishURL   *string        `json:"heartbeatFinishUrl"`
	HeartbeatErrorURL    *string        `json:"heartbeatErrorUrl"`
	StatusBadgeURL       string         `json:"statusBadgeUrl"`
	StatusBadgeJSONURL   string         `json:"statusBadgeJsonUrl"`
	UptimeBadgeURL       string         `json:"uptimeBadgeUrl"`
	UptimeBadgeJSONURL   string         `json:"uptimeBadgeJsonUrl"`
	LatencyBadgeURL      string         `json:"latencyBadgeUrl"`
	LatencyBadgeJSONURL  string         `json:"latencyBadgeJsonUrl"`
	BadgeMarkdown        string         `json:"badgeMarkdown"`
	FollowRedirects      bool           `json:"follow_redirects"`
	VerifyTLS            bool           `json:"verify_tls"`
	ProxyURL             *string        `json:"proxy_url"`
	Status               string         `json:"status"`
	RetentionDays        int64          `json:"retention_days"`
	Uptime               gqlUptime      `json:"uptime"`
	Conditions           []gqlCondition `json:"conditions"`
	Probes               []gqlID        `json:"probes"`
	NotificationChannels []gqlID        `json:"notificationChannels"`
}

type gqlCondition struct {
	Expression string `json:"expression"`
}

type gqlID struct {
	ID string `json:"id"`
}

type gqlProbe struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Queue     string `json:"queue"`
	Enabled   bool   `json:"enabled"`
	IsDefault bool   `json:"is_default"`
}

type gqlNotificationChannel struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Type   string        `json:"type"`
	Config []gqlKeyValue `json:"config"`
}

type gqlStatusPageListing struct {
	PublicName *string `json:"public_name"`
	Monitor    gqlID   `json:"monitor"`
}

type gqlStatusPage struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Slug              string                 `json:"slug"`
	CustomDomain      *string                `json:"custom_domain"`
	Headline          *string                `json:"headline"`
	Description       *string                `json:"description"`
	LogoURL           *string                `json:"logo_url"`
	FaviconURL        *string                `json:"favicon_url"`
	FooterText        *string                `json:"footer_text"`
	CustomCSS         *string                `json:"custom_css"`
	Theme             string                 `json:"theme"`
	Published         bool                   `json:"published"`
	ShowTargets       bool                   `json:"show_targets"`
	PasswordProtected bool                   `json:"passwordProtected"`
	RefreshSeconds    int64                  `json:"refresh_seconds"`
	PathURL           string                 `json:"pathUrl"`
	PublicURL         string                 `json:"publicUrl"`
	Health            string                 `json:"health"`
	Listings          []gqlStatusPageListing `json:"listings"`
}

type gqlMaintenanceWindow struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Message      *string `json:"message"`
	StartsAt     string  `json:"starts_at"`
	EndsAt       *string `json:"ends_at"`
	AppliesToAll bool    `json:"applies_to_all"`
	Phase        string  `json:"phase"`
	Monitors     []gqlID `json:"monitors"`
}
