package tftp

import (
	"encoding/binary"
	"fmt"
)

// ReadFrame returns an unpacked TFTP message from its packed representation.
func (f *Framer) ReadFrame() (Frame, error) {
	var op OpcodeType
	if err := binary.Read(f, binary.BigEndian, &op); err != nil {
		return nil, err
	}
	return f.parseFrame(op)
}

func (f *Framer) parseFrame(op OpcodeType) (Frame, error) {
	header := FrameHdr{op}
	frame, err := newFrame(op)
	if err != nil {
		return nil, err
	}
	if err = frame.read(header, f); err != nil {
		return nil, err
	}
	return frame, nil
}

func newFrame(op OpcodeType) (Frame, error) {
	ctor, ok := frameCtor[op]
	if !ok {
		return nil, fmt.Errorf("Unknown opcode %d", op)
	}
	return ctor(), nil
}

var frameCtor = map[OpcodeType]func() Frame{
	RRQ:   func() Frame { return new(RRQFrame) },
	WRQ:   func() Frame { return new(WRQFrame) },
	DATA:  func() Frame { return new(DATAFrame) },
	ACK:   func() Frame { return new(ACKFrame) },
	ERROR: func() Frame { return new(ERRORFrame) },
}

func (frame *RRQFrame) read(h FrameHdr, f *Framer) error {
	frame.Hdr = h

	filenameStr, err := f.ReadString(0x00)
	if err != nil {
		return err
	}
	frame.Filename = filenameStr[:len(filenameStr)-1]

	modeStr, err := f.ReadString(0x00)
	if err != nil {
		return err
	}
	frame.Mode = modeStr[:len(modeStr)-1]

	return nil
}

func (frame *WRQFrame) read(h FrameHdr, f *Framer) error {
	frame.Hdr = h

	filenameStr, err := f.ReadString(0x00)
	if err != nil {
		return err
	}
	frame.Filename = filenameStr[:len(filenameStr)-1]

	modeStr, err := f.ReadString(0x00)
	if err != nil {
		return err
	}
	frame.Mode = modeStr[:len(modeStr)-1]

	return nil
}

func (frame *DATAFrame) read(h FrameHdr, f *Framer) error {
	frame.Hdr = h

	if err := binary.Read(f, binary.BigEndian, &frame.BlockNum); err != nil {
		return err
	}

	frame.Data = f.Bytes()
	return nil
}

func (frame *ACKFrame) read(h FrameHdr, f *Framer) error {
	frame.Hdr = h

	if err := binary.Read(f, binary.BigEndian, &frame.BlockNum); err != nil {
		return err
	}
	return nil
}

func (frame *ERRORFrame) read(h FrameHdr, f *Framer) error {
	frame.Hdr = h

	if err := binary.Read(f, binary.BigEndian, &frame.ErrorCode); err != nil {
		return err
	}

	errStr, err := f.ReadString(0x00)
	if err != nil {
		return err
	}
	frame.ErrMsg = errStr[:len(errStr)-1]

	return nil
}
