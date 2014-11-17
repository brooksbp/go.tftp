package tftp

import (
	"log"
	"net"
	"os"
)

type Client struct {
	hostAddr *net.UDPAddr
}

func NewClient(hostAddr string) (*Client, error) {
	// Save the host addr
	addr, err := net.ResolveUDPAddr("udp", hostAddr)
	if err != nil {
		return nil, err
	}
	return &Client{
		hostAddr: addr,
	}, nil
}

// Get fname from server and store in file.
func (c *Client) Get(fname string, file string) error {

	return nil
}

// Put file on server as fname.
func (c *Client) Put(fKey string, fName string) error {

	// Get a local address and setup a listening conn on it.
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	// Send initial WRQ.
	msg := &WRQFrame{
		Filename: fKey,
		Mode:     "octet",
	}
	if err := sendFrame(conn, c.hostAddr, msg); err != nil {
		return err
	}

	// Verify ACK.

	// Open local file for writing to server.
	file, err := os.Open(fName)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Print("Close")
		}
	}()

	return nil
}
