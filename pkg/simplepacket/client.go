package simplepacket

import "net"

type Client struct {
	Connection net.Conn
	ID         uint32
	Connected  bool
}
