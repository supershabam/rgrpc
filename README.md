# rgrpc

rgrpc (remote gRPC) allows a remote behind-the-firewall device to provice a
gRPC server for exposing functionality back to a dialable gRPC client.

This is achieved by flipping the roles of server and client.
With rgrpc, the server initiates the network connection to the "client."
Once communication is established, the client is able to interact with
the server via gRPC like normal.

# license

This software is free software under the Mozilla Public License Version 2.0
