package eusend

import (
	"net/url"
	"strconv"
)

func itoa(n int) string { return strconv.Itoa(n) }

// queryString builds a "?a=b&c=d" suffix, omitting empty values. An empty map
// yields "".
func queryString(params map[string]string) string {
	q := url.Values{}
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}
