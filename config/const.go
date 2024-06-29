package config

const (
	DEFAULT_NET      = UDP_NET
	DEFAULT_TAG      = "default"
	DEFAULT_PROTOCOL = DNS_PROTOCOL
)

var (
	DEFAULT_CACHE_SIZE                                     = int32(10240)
	DEFAULT_CACHEEXPIRED_REPLY_TTL                         = int64(5)
	DEFAULT_CACHEEXPIRED_REPLY_TTL_MULTIPREFETCHSPEEDCHECK = int64(15)
	DEFAULT_CACHEEXPIRED_PREFETCH_TIMESECOND               = int64(28800)
	DEFAULT_DUALSTACK_IP_SELECTION_THRESHOLD               = int64(10)
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
	HTTP_SPEED_CHECK_TYPE SpeedCheckType = "http"
	TCP_SPEED_CHECK_TYPE  SpeedCheckType = "tcp"
	// At present, there are not many HTTP3, but it will be implemented in the future.
	// UDP_SPEED_CHECK_TYPE  SpeedCheckType = "udp"
)

type SpeedCheckType string
