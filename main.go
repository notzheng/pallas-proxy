package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		modifyRequest(req)
	}

	proxy.ModifyResponse = modifyResponse()

	return proxy, nil

}

func modifyRequest(req *http.Request) {
	req.Host = "gdmf.apple.com"
}

func modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		mimeType, _, err := mime.ParseMediaType(resp.Header.Get("content-type"))
		if err != nil {
			return err
		}
		if mimeType != "application/json" {
			return err
		}
		bodyByte, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = resp.Body.Close()
		if err != nil {
			return err
		}
		bodyString := string(bodyByte)
		if !strings.HasPrefix(bodyString, "e") {
			resp.Body = io.NopCloser(bytes.NewReader(bodyByte))
			return nil
		}
		jwtBody := strings.Split(bodyString, ".")[1]

		decodedJwtBody, err := base64.RawStdEncoding.DecodeString(jwtBody)
		if err != nil {
			return err
		}
		body := io.NopCloser(bytes.NewReader(decodedJwtBody))
		resp.Body = body
		resp.ContentLength = int64(len(decodedJwtBody))
		resp.Header.Set("Content-Length", strconv.Itoa(len(decodedJwtBody)))
		return nil
	}
}

func ProxyRequestHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	proxy, err := NewProxy("https://gdmf.apple.com")

	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", ProxyRequestHandler(proxy))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
