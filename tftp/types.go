package tftp

import (
	"bytes"
)

// OpcodeType stores the opcode field in a TFTP message.
type OpcodeType uint16

const (
	RRQ   = 1
	WRQ   = 2
	DATA  = 3
	ACK   = 4
	ERROR = 5
)

const (
	MaxDataSize = 512
	MaxMsgSize  = 2 + 2 + MaxDataSize
)

const (
	ErrorNotDefined        = 0
	ErrorFileNotFound      = 1
	ErrorAccessViolation   = 2
	ErrorDiskFull          = 3
	ErrorIllegalOp         = 4
	ErrorUnknownTid        = 5
	ErrorFileAlreadyExists = 6
	ErrorNoSuchUser        = 7
)

const (
	ModeNetascii = "netascii"
	ModeOctet    = "octet"
	ModeMail     = "mail"
)

// Frame contains all the fields in a TFTP message in an unpacked
// representation.
type Frame interface {
	write(f *Framer) error
	read(h FrameHdr, f *Framer) error
}

type FrameHdr struct {
	Opcode OpcodeType
}
type RRQFrame struct {
	Hdr      FrameHdr
	Filename string
	Mode     string
}
type WRQFrame struct {
	Hdr      FrameHdr
	Filename string
	Mode     string
}
type DATAFrame struct {
	Hdr      FrameHdr
	BlockNum uint16
	Data     []byte
}
type ACKFrame struct {
	Hdr      FrameHdr
	BlockNum uint16
}
type ERRORFrame struct {
	Hdr       FrameHdr
	ErrorCode uint16
	ErrMsg    string
}

// Framer handles serializing/deserializing TFTP messages.
type Framer struct {
	*bytes.Buffer
}

func NewFramer(b *bytes.Buffer) *Framer {
	return &Framer{b}
}
