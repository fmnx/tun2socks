package argo

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/xjasonlyu/tun2socks/v2/dialer"
	"net"
	"net/http"
	"strings"
)

type Websocket struct {
	headers  http.Header
	cdnIP    string
	Url      string
	Scheme   string
	address  string
	wsDialer *websocket.Dialer
}

func NewWebsocket(scheme, cdnIP, port, host, path string) *Websocket {

	wsDialer := &websocket.Dialer{
		TLSClientConfig: nil,
		Proxy:           http.ProxyFromEnvironment,
	}

	address := fmt.Sprintf("%s:%s", cdnIP, port)
	if strings.Contains(cdnIP, ":") {
		address = fmt.Sprintf("[%s]:%s", cdnIP, port)
	}
	wsDialer.NetDial = func(network, addr string) (net.Conn, error) {
		if cdnIP != "" {
			return dialer.Dial(network, address)
		}
		return dialer.Dial(network, addr)
	}

	headers := make(http.Header)
	headers.Set("Host", host)
	headers.Set("User-Agent", "DEV")

	return &Websocket{
		wsDialer: wsDialer,
		headers:  headers,
		cdnIP:    cdnIP,
		Scheme:   scheme,
		address:  address,
		Url:      fmt.Sprintf("%s://%s%s", scheme, host, path),
	}

}

func (w *Websocket) getDialer(ctx context.Context) *websocket.Dialer {
	wsDialer := &websocket.Dialer{}
	wsDialer.NetDial = func(network, addr string) (net.Conn, error) {
		if w.cdnIP != "" {
			return dialer.DialContext(ctx, network, w.address)
		}
		return dialer.DialContext(ctx, network, addr)
	}
	return wsDialer
}

func (w *Websocket) getHeaders(network, address string) http.Header {
	dst := make(http.Header)
	for k, v := range w.headers {
		dst[k] = v
	}
	dst.Set("Forward-Dest", address)
	dst.Set("Forward-Proto", network)
	return dst
}

func (w *Websocket) CreateWebsocketStream(ctx context.Context, network, address string) (net.Conn, error) {
	wsDialer := w.wsDialer
	if ctx != nil {
		wsDialer = w.getDialer(ctx)
	}
	wsConn, resp, err := wsDialer.Dial(w.Url, w.getHeaders(network, address))

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	return &GorillaConn{Conn: wsConn}, nil
}
