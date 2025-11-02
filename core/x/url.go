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

func IsMatchingURI(uri *url.URL, haystack []string) bool {
	for _, target := range haystack {
		if target == uri.String() {
			return true
		} else if isMatchingAsLoopback(uri, target) {
			// loopback address can be seen as matching with different ports
			return true
		}
	}

	return false
}

func isMatchingAsLoopback(uri *url.URL, target string) bool {
	registered, err := url.Parse(target)
	if err != nil {
		return false
	}

	if uri.Scheme == "http" &&
		IsLoopback(uri.Hostname()) &&
		uri.Hostname() == registered.Hostname() &&
		uri.Path == registered.Path &&
		uri.RawQuery == uri.RawQuery {
		return true
	}

	return false
}
