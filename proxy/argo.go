package proxy

import (
	"context"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy/proto"
	"github.com/xjasonlyu/tun2socks/v2/transport/argo"
	"net"
)

var _ Proxy = (*Argo)(nil)

type Argo struct {
	*Base
	ws *argo.Websocket
}

func (a *Argo) Proto() proto.Proto {
	return a.Base.Proto()
}

func (a *Argo) Addr() string {
	return a.ws.Url
}

func NewArgo(ws *argo.Websocket) *Argo {
	return &Argo{
		Base: &Base{
			proto: proto.Argo,
		},
		ws: ws,
	}
}

func (a *Argo) DialContext(ctx context.Context, metadata *M.Metadata) (net.Conn, error) {
	c, err := a.ws.CreateWebsocketStream(ctx, "tcp", metadata.DestinationAddress())
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (a *Argo) DialUDP(metadata *M.Metadata) (net.PacketConn, error) {
	c, err := a.ws.CreateWebsocketStream(nil, "udp", metadata.DestinationAddress())
	if err != nil {
		return nil, err
	}
	return &argoPacketConn{Conn: c}, nil
}

type argoPacketConn struct {
	net.Conn
	rAddr net.Addr
}

func (w argoPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, err = w.Conn.Read(p)
	if err != nil {
		return 0, nil, err
	}
	return n, w.rAddr, nil
}

func (w argoPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	n, err = w.Conn.Write(p)
	if err != nil {
		return 0, err
	}
	w.rAddr = addr
	return n, nil
}
