package client

import "net"

type Client struct {
	CID     string
	IPddr   string
	UDPAddr *net.UDPAddr
}
