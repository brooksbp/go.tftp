package tftp

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

func addrEqual(a, b *net.UDPAddr) bool {
	return a.IP.Equal(b.IP) && a.Port == b.Port && a.Zone == b.Zone
}

func resolveAndListen(addrStr string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		log.Print("ResolveUDPAddr:", err)
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Print("ListenUDP:", err)
		return nil, err
	}
	return conn, nil
}

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
