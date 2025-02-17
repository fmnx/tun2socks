package argo

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/xjasonlyu/tun2socks/v2/dialer"
	"net"
	"net/http"
)

type Websocket struct {
	headers  http.Header
	cdnIP    string
	port     string
	url      string
	wsDialer *websocket.Dialer
}

func NewWebsocket(scheme, cdnIP, port, host, path string) *Websocket {

	wsDialer := &websocket.Dialer{
		TLSClientConfig: nil,
		Proxy:           http.ProxyFromEnvironment,
	}
	wsDialer.NetDial = func(network, addr string) (net.Conn, error) {
		if cdnIP != "" {
			return dialer.Dial(network, fmt.Sprintf("%s:%s", cdnIP, port))
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
		port:     port,
		url:      fmt.Sprintf("%s://%s%s", scheme, host, path),
	}

}

func (w *Websocket) getDialer(ctx context.Context) *websocket.Dialer {
	wsDialer := &websocket.Dialer{}
	wsDialer.NetDial = func(network, addr string) (net.Conn, error) {
		// 连接指定的 IP 地址而不是解析域名
		if w.cdnIP != "" {
			return dialer.DialContext(ctx, network, fmt.Sprintf("%s:%d", w.cdnIP, w.port))
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
	wsConn, resp, err := wsDialer.Dial(w.url, w.getHeaders(network, address))

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	return &GorillaConn{Conn: wsConn}, nil
}
