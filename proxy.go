package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const dirtyRedirect string = "<script>window.location.replace('/');</script>"

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

	if cookie == nil {
		return false
	}

	return VerifyToken(cookie.Value, "#INDRAK#")
}

func proxyhandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/jwaax_authenticate" {
		jt := r.URL.Query().Get("jt")
		if jt == "" {
			//println("404")
			w.WriteHeader(404)
		} else {
			//println("Got Cookie")
			//println(jt)
			cookie := &http.Cookie{
				Name:    "jt",
				Value:   jt,
				Expires: time.Now().Add(time.Hour * 24 * 30),
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/jwaax_redir", http.StatusFound)

			return
		}
	} else if r.URL.Path == "/jwaax_redir" {
		fmt.Fprintf(w, "Hello!")
		return
	}

	url := url.URL{
		Scheme: "http",
		Host:   "www.example.com",
	}

	if isauthed(r) {
		proxy := NewMultipleHostReverseProxy(&url)
		proxy.ServeHTTP(w, r)
	} else {
		requestToken := GenerateToken("0", "beyond.demille.tech/jwaax_authenticate")
		http.Redirect(w, r, "http://secure.demilletech.net/external/signin/?request_token="+requestToken, http.StatusFound)
	}
}

func main() {
	http.HandleFunc("/", proxyhandler)
	println("Starting Server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
