package config

const (
	DEFAULT_NET      = UDP_NET
	DEFAULT_TAG      = "default"
	DEFAULT_PROTOCOL = DNS_PROTOCOL
)

type Protocol string

const (
	DNS_PROTOCOL   Protocol = "dns"
	SOCK5_PROTOCOL Protocol = "sock5"
	HTTPS_PROTOCOL Protocol = "https"
)

type Net string

const (
	UDP_NET Net = "udp"
	TCP_NET Net = "tcp"
	TLS_NET Net = "tcp-tls"
)

type CacheMissResponseMode string

const (
	// The fastest outbound dns + ping response mode, DNS query delay + ping delay is the shortest;
	// may take some time to test speed for first response upstream's ips.
	FIRST_PING_RESPONSEMODE CacheMissResponseMode = "first-ping"
	// The fastest IP address mode, return the fastest ip address
	// may take some time to test speed for all upstreams's ips.
	FASTEST_IP_RESPONSEMODE CacheMissResponseMode = "fastest-ip"
	// The fastest response DNS result mode, the DNS query waiting time is the shortest.
	// not wait any test speed
	FASTEST_RESPONSE_RESPONSEMODE CacheMissResponseMode = "fastest-response"
)

const (
	PING_SPEED_CHECK_TYPE SpeedCheckType = "ping"
	TCP_SPEED_CHECK_TYPE  SpeedCheckType = "tcp"
	// At present, there are not many HTTP3, but it will be implemented in the future.
	// UDP_SPEED_CHECK_TYPE  SpeedCheckType = "udp"
)

type SpeedCheckType string
