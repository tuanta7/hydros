package x

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

func IsValidRedirectURI(redirectURI string) bool {
	ru, err := url.Parse(redirectURI)
	if err != nil {
		return false
	}

	if ru == nil {
		return false
	}

	if ru.Scheme == "" {
		return false
	}

	if ru.Fragment != "" {
		return false
	}

	return true
}

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

func MatchRedirectURI(uri string, haystack []string) (*url.URL, error) {
	if len(haystack) == 1 && uri == "" {
		parsed, err := url.Parse(haystack[0])
		if err != nil {
			return nil, err
		}

		return parsed, nil
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	for _, target := range haystack {
		if target == uri {
			return parsed, nil
		} else if isMatchingAsLoopback(parsed, target) {
			// loopback address can be seen as matching with different ports
			return parsed, nil
		}
	}

	return nil, errors.New("no matching redirect uri found")
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
