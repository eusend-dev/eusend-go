package eusend

import (
	"net/url"
	"strconv"
)

func itoa(n int) string { return strconv.Itoa(n) }

// queryString builds a "?a=b&c=d" suffix, omitting empty values. An empty map
// yields "".
func queryString(params map[string]string) string {
	return queryStringWith(params, nil)
}

// queryStringWith is queryString plus repeatable params (e.g. "?tag=a&tag=b"),
// which the tag filter on /emails uses to AND several filters together.
func queryStringWith(params map[string]string, repeated map[string][]string) string {
	q := url.Values{}
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	for k, values := range repeated {
		for _, v := range values {
			if v != "" {
				q.Add(k, v)
			}
		}
	}
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}
