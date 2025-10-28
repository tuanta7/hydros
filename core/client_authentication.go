package core

import (
	"context"
	"net/http"
	"net/url"
)

const (
	keyClientAssertion     = "client_assertion"
	keyClientAssertionType = "client_assertion_type"
	clientAssertionJWTType = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
)

// AuthenticateClient check if the client needs authentication and returns the client if it does.
func (o *OAuth2) AuthenticateClient(ctx context.Context, r *http.Request, form url.Values) (Client, error) {
	if at := form.Get(keyClientAssertionType); at == clientAssertionJWTType {
		assertion := form.Get(keyClientAssertion)
		if len(assertion) == 0 {
			return nil, ErrInvalidRequest.WithHint("Missing client assertion")
		}

		// TODO: implement rfc7521 client assertion
	} else if len(at) > 0 {
		return nil, ErrInvalidRequest.WithHint("Unsupported client assertion type: " + at)
	}

	clientID, clientSecret, err := clientCredentialsFromRequest(r, form)
	if err != nil {
		return nil, err
	}

	client, err := o.store.GetClient(ctx, clientID)
	if err != nil {
		return nil, ErrInvalidClient.WithWrap(err).WithDebug(err.Error())
	}

	// Skip authentication for public clients
	if client.IsPublic() {
		return client, nil
	}

	err = o.config.GetSecretsHasher().Compare(ctx, client.GetHashedSecret(), []byte(clientSecret))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func clientCredentialsFromRequest(r *http.Request, form url.Values) (clientID, clientSecret string, err error) {
	if id, secret, ok := r.BasicAuth(); ok {
		clientID, err = url.QueryUnescape(id)
		if err != nil {
			return "", "", ErrInvalidRequest.WithHint("The client id in the HTTP authorization header could not be decoded from 'application/x-www-form-urlencoded'.").WithWrap(err).WithDebug(err.Error())
		}

		clientSecret, err = url.QueryUnescape(secret)
		if err != nil {
			return "", "", ErrInvalidRequest.WithHint("The client secret in the HTTP authorization header could not be decoded from 'application/x-www-form-urlencoded'.").WithWrap(err).WithDebug(err.Error())
		}

		return clientID, clientSecret, nil
	}

	// Credentials missing in HTTP Authorization header. Try to read them from the request body.
	clientID = form.Get("client_id")
	if clientID == "" {
		return "", "", ErrInvalidRequest.WithHint("Client credentials missing or malformed in both HTTP Authorization header and HTTP POST body.")
	}

	clientSecret = form.Get("client_secret")
	return clientID, clientSecret, nil
}
