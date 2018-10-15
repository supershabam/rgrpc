package rgrpc

// raddr implements net.Addr and is returned by listener
type raddr struct {
}

// name of the network (for example, "tcp", "udp")
func (a *raddr) Network() string {
	return "rgrpc"
}

// string form of address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
func (a *raddr) String() string {
	return "rgrpc"
}
