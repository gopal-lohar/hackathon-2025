package api

import (
	"io"
	"net"

	"github.com/gopal-lohar/hackathon-2025/internal/shared/protocol"
	"github.com/gopal-lohar/hackathon-2025/internal/shared/utils"
)

func (as *APIServer) ListenForEndpointConnections() {
	as.logger.Info("Listening for endpoint connections")
	l, err := net.Listen("tcp", ":8081")
	if err != nil {
		as.logger.Fatal(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			as.logger.Error(err)
			continue
		}

		remoteAddr := conn.RemoteAddr().String()
		addr, port, err := net.SplitHostPort(remoteAddr)
		if err != nil {
			as.logger.Warnf("Error splitting host and port: %v", err)
			return
		}
		as.logger.Infof("Accepted a new connection from endpoint: %s", remoteAddr)

		id, err := as.endpointStore.CreateEndpoint(addr, port)
		if err != nil {
			as.logger.Warnf("Error creating endpoint: %v", err)
		}

		as.logger.Infof("An endpoint with id: %s has been connected successfully", id)
		as.endpointMap[id] = conn

		go as.handleEndpointConnection(conn, id)
	}
}

func (as *APIServer) handleEndpointConnection(conn net.Conn, id string) {
	// Close the connection and delete the endpoint on func end
	defer func() {
		conn.Close()
		as.logger.Infof("Connection with endpoint id: %s closed", id)

		delete(as.endpointMap, id)

		if err := as.endpointStore.DeleteEndpoint(id); err != nil {
			as.logger.Warnf("Error deleting endpoint with id %s: %v", id, err)
		} else {
			as.logger.Infof("Endpoint with id: %s has been removed from database", id)
		}
	}()

	// Read loop
	for {
		netMsg, err := utils.ReceiveNetMsg(conn)
		if err != nil {
			if err == io.EOF {
				as.logger.Infof("Connection with endpoint id: %s was closed by client", id)
			} else {
				as.logger.Warnf("Error reading from connection with endpoint id: %s: %v", id, err)
			}
			return
		}
		// TODO: Handle messages here
		as.logger.Infof("Received message from endpoint id: %s: %s", id, netMsg)
		as.handleEndpointMsg(netMsg)
	}
}

func (as *APIServer) handleEndpointMsg(netMsg *protocol.NetworkMessage) {
}
