package x

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// BindForm reads form data from both application/x-www-form-urlencoded and multipart/form-data.
func BindForm(r *http.Request) (url.Values, error) {
	err := r.ParseMultipartForm(1 << 20)
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return nil, err
	}

	return r.Form, nil
}

// BindPostForm is the same as BindForm, but return post form data only.
func BindPostForm(r *http.Request) (url.Values, error) {
	err := r.ParseMultipartForm(1 << 20)
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return nil, err
	}

	return r.PostForm, nil
}

func AccessTokenFromRequest(r *http.Request) string {
	parts := strings.SplitN(r.Header.Get("Authorization"), " ", 3)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		form, err := BindForm(r) // this gets both url and body params
		if err != nil {
			return ""
		}

		return form.Get("access_token")
	}

	return parts[1]
}

func EscapeJSONString(str string) string {
	// Escape reverse solidus.
	str = strings.ReplaceAll(str, `\`, `\\`)
	// Escape control characters.
	for r := rune(0); r < ' '; r++ {
		str = strings.ReplaceAll(str, string(r), fmt.Sprintf(`\u%04x`, r))
	}
	// Escape quotation mark.
	str = strings.ReplaceAll(str, `"`, `\"`)
	return str
}
