package urlx

import "net/url"

func Copy(src *url.URL) *url.URL {
	var out = new(url.URL)
	*out = *src
	return out
}

func AppendQueryString(u string, query url.Values) (string, error) {
	pu, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	return AppendQuery(pu, query).String(), nil
}

func AppendQuery(u *url.URL, query url.Values) *url.URL {
	ep := Copy(u)
	q := ep.Query()

	for k := range query {
		q.Set(k, query.Get(k))
	}

	ep.RawQuery = q.Encode()
	return ep
}
