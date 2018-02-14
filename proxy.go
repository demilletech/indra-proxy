package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// NewMultipleHostReverseProxy creates a reverse proxy that will randomly
// select a host from the passed `targets`
func NewMultipleHostReverseProxy(target *url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		//println("CALLING DIRECTOR")
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
	}

	proxy := func(req *http.Request) (*url.URL, error) {
		//println("CALLING PROXY")
		return http.ProxyFromEnvironment(req)
	}

	dial := func(network, addr string) (net.Conn, error) {
		//println("CALLING DIAL")
		conn, err := (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial(network, addr)
		if err != nil {
			//println("Error during DIAL:", err.Error())
		}
		return conn, err
	}
	return &httputil.ReverseProxy{
		Director: director,
		Transport: &http.Transport{
			Proxy:               proxy,
			Dial:                dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func isauthed(r *http.Request) bool {
	cookie, _ := r.Cookie("jt")
	//println(cookie.Value)

	if cookie == nil {
		return false
	}

	resp, err := http.Get("https://secure.demilletech.net/api/key")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	keyb, _ := ioutil.ReadAll(resp.Body)
	key := string(keyb)

	println(key)

	return true
}

func proxyhandler(w http.ResponseWriter, r *http.Request) {
	/*proxy := NewMultipleHostReverseProxy([]*url.URL{
		{
			Scheme: "http",
			Host:   "127.0.0.1:5000",
		},
	})*/

	//url, _ := url.Parse("http://example.com/")
	url := url.URL{
		Scheme: "http",
		Host:   "www.example.com",
	}

	if isauthed(r) {
		println("Authed!")
	} else {
		println("Unauthed!")
	}

	proxy := NewMultipleHostReverseProxy(&url)
	proxy.ServeHTTP(w, r)

}

func main() {
	http.HandleFunc("/", proxyhandler)
	println("Starting Server")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
