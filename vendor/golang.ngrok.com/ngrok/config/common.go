package config

type commonOpts struct {
	// Restrictions placed on the origin of incoming connections to the edge.
	CIDRRestrictions *cidrRestrictions
	// The version of PROXY protocol to use with this tunnel, zero if not
	// using.
	ProxyProto ProxyProtoVersion
	// Tunnel-specific opaque metadata. Viewable via the API.
	Metadata string
	// Tunnel backend metadata. Viewable via the dashboard and API, but has no
	// bearing on tunnel behavior.
	// If not set, defaults to a URI in the format `app://hostname/path/to/executable?pid=12345`
	ForwardsTo string
}

func (cfg *commonOpts) getForwardsTo() string {
	if cfg.ForwardsTo == "" {
		return defaultForwardsTo()
	}
	return cfg.ForwardsTo
}
