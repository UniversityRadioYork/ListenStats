package cdns

import (
	"io"
	"net/http"
	"strings"
)

func getUrlAndParseLines(url string) ([]string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	buf := new(strings.Builder)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return nil, err
	}
	values := strings.Split(buf.String(), "\n")
	return values, nil
}

func GetCloudflareIPRanges() ([]string, error) {
	rangesV4, err := getUrlAndParseLines("https://www.cloudflare.com/ips-v4")
	if err != nil {
		return nil, err
	}
	rangesV6, err := getUrlAndParseLines("https://www.cloudflare.com/ips-v6")
	if err != nil {
		return nil, err
	}
	return append(rangesV4, rangesV6...), nil
}

func GetProjectShieldIPRanges() ([]string, error) {
	// So far PS only has one IP. Yeet.
	return []string{"35.235.224.0/20"}, nil
}
