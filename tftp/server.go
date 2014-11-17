package tftp

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// Server is an TFTP server.
type Server struct {
	conn *net.UDPConn

	// TODO: decouple storage from server.
	mu sync.RWMutex
	m  map[string][]byte
}

// fetchFile returns a reader interface to file referenced by |fname|.
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

// putFile stores file |fname|.
func (s *Server) putFile(fname *string, f *bytes.Buffer) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.m == nil {
		s.m = make(map[string][]byte)
	} else {
		// Verify that the file does not already exist.
		_, ok := s.m[*fname]
		if ok {
			return false
		}
	}

	s.m[*fname] = f.Bytes()
	return true
}

func NewServer(listenAddr string) (*Server, error) {
	conn, err := resolveAndListen(listenAddr)
	if err != nil {
		return nil, err
	}
	return &Server{conn: conn}, nil
}

func (s *Server) Run() error {
	var buf [MaxMsgSize]byte

	for {
		n, caddr, err := s.conn.ReadFromUDP(buf[0:])
		if err != nil {
			return err
		}

		framer := NewFramer(bytes.NewBuffer(buf[:n]))
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
	// TODO: s.conn.Close()
}

// Recieve a DATA msg from |peerAddr| and write contents into |writer|. Reply
// with an ACK msg containing the newely received blocknum.
func recvData(conn *net.UDPConn, peerAddr *net.UDPAddr, writer *bytes.Buffer) (int, error) {
	buf := make([]byte, MaxMsgSize)
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return 0, err
	}
	if !addrEqual(addr, peerAddr) {
		return 0, fmt.Errorf("Expecting addr=%s, got=%s", peerAddr, addr)
	}

	framer := NewFramer(bytes.NewBuffer(buf[:n]))
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

// Send DATA message |dataFrame| to client |peerAddr|.
func sendData(conn *net.UDPConn, peerAddr *net.UDPAddr, dataFrame *DATAFrame) error {
	blockNum := (*dataFrame).BlockNum

	framer := NewFramer(new(bytes.Buffer))

	// TODO: handle timeouts, retransmission, blknum ordering issues..

	if err := framer.WriteFrame(dataFrame); err != nil {
		return err
	}
	if _, err := conn.WriteToUDP(framer.Bytes(), peerAddr); err != nil {
		return err
	}
	framer.Reset()

	buf := make([]byte, MaxMsgSize)
	_, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return err
	}
	if !addrEqual(addr, peerAddr) {
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

// handleReadRequest handles a new incoming RRQ to the server. Setup an
// ephemeral port to use for the new connection and serve the file.
func (s *Server) handleReadRequest(peerAddr *net.UDPAddr, rrqFrame *RRQFrame) {
	var fname string = rrqFrame.Filename
	var mode string = rrqFrame.Mode

	conn, err := resolveAndListen(":0")
	if err != nil {
		return
	}

	if mode != ModeOctet {
		log.Print("Unhandled mode:", mode)
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

	// Read file MaxDataSize bytes at a time and send to client.
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

	if err := conn.Close(); err != nil {
		log.Print(err)
	}
}

// handleWriteRequest handles a new incoming WRQ to the server. Setup an
// emphemeral port to use for the new connection and recieve incoming data.
func (s *Server) handleWriteRequest(peerAddr *net.UDPAddr, wrqFrame *WRQFrame) {
	var fname string = wrqFrame.Filename
	var mode string = wrqFrame.Mode

	conn, err := resolveAndListen(":0")
	if err != nil {
		return
	}

	if mode != ModeOctet {
		log.Print("Unhandled mode:", mode)
		return
	}

	// If file exists, send an error message.
	if _, exists := s.fetchFile(&fname); exists {
		msg := &ERRORFrame{
			ErrorCode: uint16(ErrorFileAlreadyExists),
			ErrMsg:    "",
		}
		if err := sendFrame(conn, peerAddr, msg); err != nil {
			return
		}
		return
	}

	// Acknowledge the write request with an ACK BlockNum=0 message.
	msg := &ACKFrame{
		BlockNum: uint16(0),
	}
	if err := sendFrame(conn, peerAddr, msg); err != nil {
		return
	}

	// Receive DATA messages containing |fname| data.
	writer := new(bytes.Buffer)

	for {
		n, err := recvData(conn, peerAddr, writer)
		if err != nil {
			log.Println(err)
			return
		}
		if n < MaxDataSize {
			break
		}
	}

	// Store |fname|.
	if !s.putFile(&fname, writer) {
		return
	}

	if err := conn.Close(); err != nil {
		log.Print(err)
	}

}
