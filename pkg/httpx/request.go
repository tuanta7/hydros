package httpx

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

func ReadForm(r *http.Request) (url.Values, error) {
	err := r.ParseMultipartForm(1 << 20)
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return nil, err
	}

	return r.Form, nil
}

func AccessTokenFromRequest(r *http.Request) string {
	parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		form, err := ReadForm(r)
		if err != nil {
			return ""
		}

		return form.Get("access_token")
	}

	return parts[1]
}

