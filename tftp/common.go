package tftp

import (
	"bytes"
	"fmt"
	"net"
)

// Pack frame and send it to peerAddr.
func sendFrame(conn *net.UDPConn, peerAddr *net.UDPAddr, frame Frame) error {
	buffer := new(bytes.Buffer)
	framer := NewFramer(buffer)

	if err := framer.WriteFrame(frame); err != nil {
		return fmt.Errorf("WriteFrame")
	}
	if _, err := conn.WriteToUDP(framer.Bytes(), peerAddr); err != nil {
		return err
	}
	return nil
}
