package tftp

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Server struct {
	listenAddr *net.UDPAddr
	conn       *net.UDPConn

	mu sync.RWMutex
	m  map[string][]byte
}

func NewServer(listenAddr string) (*Server, error) {
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	return &Server{
		listenAddr: addr,
		conn:       conn,
	}, nil
}

func (s *Server) Run() error {
	var buf [MaxMsgSize]byte

	for {
		n, caddr, err := s.conn.ReadFromUDP(buf[0:])
		if err != nil {
			return err
		}

		framer := NewFramer(bytes.NewBuffer(buf[0:n]))

		frame, err := framer.ReadFrame()
		if err != nil {
			log.Print("ReadFrame:", err)
			continue
		}

		switch frame := frame.(type) {
		case *RRQFrame:
			go s.handleReadRequest(caddr, frame)
		case *WRQFrame:
			go s.handleWriteRequest(caddr, frame)
		default:
			log.Print("Unhandled conn:", frame)
		}
	}
}

func newConnection() (*net.UDPConn, error) {
	saddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		log.Print("ResolveUDPAddr:", err)
		return nil, err
	}
	conn, err := net.ListenUDP("udp", saddr)
	if err != nil {
		log.Print("ListenUDP:", err)
		return nil, err
	}
	return conn, nil
}

// If |fname| exists, return a reader interface and True, otherwise nil.
func (s *Server) fetchFile(fname *string) (*bytes.Reader, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.m == nil {
		s.m = make(map[string][]byte)
		return nil, false
	}

	_, ok := s.m[*fname]
	if !ok {
		return nil, false
	}

	return bytes.NewReader(s.m[*fname]), true
}

func (s *Server) putFile(fname *string, f *bytes.Buffer) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.m == nil {
		s.m = make(map[string][]byte)
	} else {
		// Verify again that the file does not exist.
		_, ok := s.m[*fname]
		if ok {
			return false
		}
	}

	s.m[*fname] = f.Bytes()
	return true
}

func UDPAddrEqual(a, b *net.UDPAddr) bool {
	return a.IP.Equal(b.IP) && a.Port == b.Port && a.Zone == b.Zone
}

// Recieve a DATA msg from |peerAddr| and write contents into |writer|. Reply
// with an ACK msg containing the newely received blocknum.
func recvData(conn *net.UDPConn, peerAddr *net.UDPAddr, writer *bytes.Buffer) (int, error) {
	buf := make([]byte, MaxMsgSize)
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return 0, err
	}
	if !UDPAddrEqual(addr, peerAddr) {
		return 0, fmt.Errorf("Expecting addr=%s, got=%s",
			peerAddr, addr)
	}

	buffer := bytes.NewBuffer(buf[:n])
	framer := NewFramer(buffer)
	frame, err := framer.ReadFrame()
	if err != nil {
		return 0, err
	}
	dataFrame, ok := frame.(*DATAFrame)
	if !ok {
		return 0, fmt.Errorf("Expecting DATA msg.")
	}

	n, err = writer.Write(dataFrame.Data)
	if err != nil {
		return 0, err
	}

	msg := &ACKFrame{
		BlockNum: uint16(dataFrame.BlockNum),
	}
	if err := sendFrame(conn, peerAddr, msg); err != nil {
		return 0, err
	}
	return n, nil
}

func sendData(conn *net.UDPConn, peerAddr *net.UDPAddr, dataFrame *DATAFrame) error {
	blockNum := (*dataFrame).BlockNum
	buffer := new(bytes.Buffer)
	framer := NewFramer(buffer)

	// TODO: timeouts, resend, invalid blocknums..

	if err := framer.WriteFrame(dataFrame); err != nil {
		return err
	}
	if _, err := conn.WriteToUDP(framer.Bytes(), peerAddr); err != nil {
		return err
	}

	buf := make([]byte, MaxMsgSize)
	_, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return err
	}
	if addr != peerAddr {
		return fmt.Errorf("Expecting addr=%s", peerAddr)
	}
	if _, err := framer.Write(buf); err != nil {
		return err
	}
	frame, err := framer.ReadFrame()
	if err != nil {
		return err
	}
	ackFrame, ok := frame.(*ACKFrame)
	if !ok {
		return fmt.Errorf("Expecting ACK msg.")
	}
	if ackFrame.BlockNum != blockNum {
		return fmt.Errorf("Expecting ACK BlockNum=%d", blockNum)
	}
	return nil
}

func (s *Server) handleReadRequest(peerAddr *net.UDPAddr, rrqFrame *RRQFrame) {
	// Setup a new connection used to service read request.
	conn, err := newConnection()
	if err != nil {
		return
	}

	mode := rrqFrame.Mode
	fname := rrqFrame.Filename

	if mode != ModeOctet {
		log.Print("mode:", mode)
		return
	}

	reader, exists := s.fetchFile(&fname)

	// If file does not exist, send an error message.
	if !exists {
		msg := &ERRORFrame{
			ErrorCode: uint16(ErrorFileNotFound),
			ErrMsg:    "",
		}
		if err := sendFrame(conn, peerAddr, msg); err != nil {
			return
		}
		return
	}

	buf := make([]byte, MaxDataSize)

	for block := 1; ; block++ {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return
		}
		dataFrame := &DATAFrame{
			BlockNum: uint16(block),
			Data:     buf[:n],
		}
		err = sendData(conn, peerAddr, dataFrame)
		if err != nil {
			return
		}
		if n < MaxDataSize {
			break
		}
	}
}

func (s *Server) handleWriteRequest(peerAddr *net.UDPAddr, wrqFrame *WRQFrame) {
	conn, err := newConnection()
	if err != nil {
		return
	}

	mode := wrqFrame.Mode
	fname := wrqFrame.Filename

	if mode != ModeOctet {
		log.Print("mode:", mode)
		return
	}

	_, exists := s.fetchFile(&fname)

	// If file exists, send an error message.
	if exists {
		msg := &ERRORFrame{
			ErrorCode: uint16(ErrorFileAlreadyExists),
			ErrMsg:    "",
		}
		if err := sendFrame(conn, peerAddr, msg); err != nil {
			return
		}
		return
	}

	// Otherwise, send ACK block number 0.
	msg := &ACKFrame{BlockNum: uint16(0)}
	if err := sendFrame(conn, peerAddr, msg); err != nil {
		return
	}

	writer := new(bytes.Buffer)

	for {
		n, err := recvData(conn, peerAddr, writer)
		if err != nil {
			fmt.Println(err)
			return
		}
		if n < MaxDataSize {
			break
		}
	}

	// Store new file.
	if !s.putFile(&fname, writer) {
		return
	}
}
