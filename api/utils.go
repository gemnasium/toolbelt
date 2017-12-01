package api

import (
	"net/url"
	"strings"
)

type requestOptions struct {
	Method string
	URI    string
	Body   interface{}
	Result interface{}
}

//Returns the host name without ":" from an URL
func getHost(host_url string) (host string, err error) {
	u, err := url.Parse(host_url)
	if err != nil {
		return "", err
	}
	return strings.Split(u.Host, ":")[0], nil
}