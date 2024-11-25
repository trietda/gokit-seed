package common

import "net/url"

func MustParseUrl(urlString string) *url.URL {
	u, err := url.Parse(urlString)

	if err != nil {
		panic(err)
	}

	return u
}
