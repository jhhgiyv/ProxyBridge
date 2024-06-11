package router_struct

type NewProxyRequests struct {
	Proxy            string `json:"proxy" binding:"required"`
	NotAuth          bool   `json:"no_auth" binding:"required"`
	ProxyType        string `json:"proxy_type" binding:"required"`
	ProxyLifetimeSec int    `json:"proxy_lifetime_sec" binding:"required"`
}
