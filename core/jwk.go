package core

type JSONWebKeySet struct {
	Keys []JSONWebKey `json:"keys"`
}
type JSONWebKey struct {
	Key       any
	KeyID     string
	Algorithm string
	Use       string
}
