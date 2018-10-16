// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

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
