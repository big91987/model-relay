// Package apikey owns application credentials, generations and access grants.
package apikey

type AppKey struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Hint    string `json:"hint"`
	Enabled bool   `json:"enabled"`
	Created string `json:"created"`
}

type ControlledKey struct {
	AppKey
	Tenant      string           `json:"tenant_id"`
	Application string           `json:"application"`
	Expires     *string          `json:"expires_at"`
	Revoked     bool             `json:"revoked"`
	Revision    int64            `json:"revision"`
	Models      []string         `json:"models"`
	Diagnostics []string         `json:"diagnostic_members"`
	Generations []map[string]any `json:"generations"`
}

type Input struct {
	Name        string   `json:"name"`
	Application string   `json:"application"`
	Expires     *string  `json:"expires_at"`
	Enabled     bool     `json:"enabled"`
	Models      []string `json:"models"`
	Diagnostics []string `json:"diagnostic_members"`
	Revision    int64    `json:"expected_revision"`
}

type Transition struct {
	Hours    int   `json:"transition_hours"`
	Revision int64 `json:"expected_revision"`
}

type Operation struct{ Key, Path string }

type Result struct {
	ID, Secret     string
	AlreadyApplied bool
}
