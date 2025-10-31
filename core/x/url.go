package x

import (
	"net"
	"net/url"
	"strings"
)

func IsURISecure(uri *url.URL) bool {
	return !(uri.Scheme == "http" && !IsLocalhost(uri))
}

func IsLocalhost(uri *url.URL) bool {
	hostname := uri.Hostname()
	return hostname == "localhost" || strings.HasSuffix(hostname, ".localhost") || IsLoopback(hostname)
}

func IsLoopback(hostname string) bool {
	return net.ParseIP(hostname).IsLoopback()
}
