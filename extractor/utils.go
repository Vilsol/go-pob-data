package extractor

import (
	"crypto/tls"
	"io"
	"net/http"
)

var downloadClient *http.Client

func init() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	downloadClient = &http.Client{Transport: tr}
}

func Fetch(url string) []byte {
	resp, err := downloadClient.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return all
}
