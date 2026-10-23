package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"
)

type ProxyManager struct {
	proxies sync.Map
}

func NewProxyManager() *ProxyManager {
	return &ProxyManager{}
}

func (m *ProxyManager) Forward(
	w http.ResponseWriter,
	r *http.Request,
	targetURL string,
	pathPrefix string,
) error {
	target, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid target url: %w", err)
	}

	proxy, _ := m.proxies.LoadOrStore(targetURL, m.createProxy(target, pathPrefix))
	rp := proxy.(*httputil.ReverseProxy)

	rp.ServeHTTP(w, r)
	return nil
}

func (m *ProxyManager) createProxy(target *url.URL, pathPrefix string) *httputil.ReverseProxy {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)

			pr.SetXForwarded()

			trimmedPath := strings.TrimPrefix(pr.Out.URL.Path, pathPrefix)
			if !strings.HasPrefix(trimmedPath, "/") {
				trimmedPath = "/" + trimmedPath
			}
			pr.Out.URL.Path = trimmedPath

			if clientIP, _, err := net.SplitHostPort(pr.In.RemoteAddr); err == nil {
				pr.Out.Header.Set("X-Real-IP", clientIP)
			}
		},

		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},

		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "502 Bad Gateway: Upstream service unreachable", http.StatusBadGateway)
		},
	}

	return proxy
}
