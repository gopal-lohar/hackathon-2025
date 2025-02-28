package utils

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"net/http"

	"github.com/gopal-lohar/hackathon-2025/internal/shared/protocol"
	"google.golang.org/protobuf/proto"
)

// SendNetMsg takes in a connection and a network message
// and sends the network message over the connection
func SendNetMsg(conn net.Conn, msg *protocol.NetworkMessage) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	msgLen := uint32(len(data))
	if err := binary.Write(conn, binary.BigEndian, msgLen); err != nil {
		return err
	}

	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}

// ReceiveNetMsg reads a network message from a connection
// and returns the network message and an error
// can also return an io.EOF error
func ReceiveNetMsg(conn net.Conn) (*protocol.NetworkMessage, error) {
	var msgLen uint32
	if err := binary.Read(conn, binary.BigEndian, &msgLen); err != nil {
		return nil, err
	}
	data := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	var netMsg protocol.NetworkMessage
	if err := proto.Unmarshal(data, &netMsg); err != nil {
		return nil, err
	}
	return &netMsg, nil
}

func WriteErrorResponse(w http.ResponseWriter, msg string, status int) {
	w.WriteHeader(status)
	w.Write([]byte(msg))
}

func WriteSuccessResponse(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg))
}

func WriteJSONResponse(w http.ResponseWriter, msg any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(msg)
}
