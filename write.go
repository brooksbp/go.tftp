package tftp

import (
	"encoding/binary"
)

func (f *Framer) WriteFrame(frame Frame) error {
	return frame.write(f)
}

func writeHdr(f *Framer, h FrameHdr) error {
	if err := binary.Write(f, binary.BigEndian, h.Opcode); err != nil {
		return err
	}
	return nil
}

func writeString(f *Framer, s string) error {
	if _, err := f.WriteString(s); err != nil {
		return err
	}
	if err := f.WriteByte(0x00); err != nil {
		return err
	}
	return nil
}

func (frame *RRQFrame) write(f *Framer) (err error) {
	frame.Hdr.Opcode = RRQ

	if err = writeHdr(f, frame.Hdr); err != nil {
		return
	}
	if err = writeString(f, frame.Filename); err != nil {
		return
	}
	if err = writeString(f, frame.Mode); err != nil {
		return
	}
	return
}

func (frame *WRQFrame) write(f *Framer) (err error) {
	frame.Hdr.Opcode = WRQ

	if err = writeHdr(f, frame.Hdr); err != nil {
		return
	}
	if err = writeString(f, frame.Filename); err != nil {
		return
	}
	if err = writeString(f, frame.Mode); err != nil {
		return
	}
	return
}

func (frame *DATAFrame) write(f *Framer) (err error) {
	frame.Hdr.Opcode = DATA

	if err = writeHdr(f, frame.Hdr); err != nil {
		return
	}
	if err = binary.Write(f, binary.BigEndian, frame.BlockNum); err != nil {
		return
	}
	if _, err = f.Write(frame.Data); err != nil {
		return
	}
	return
}

func (frame *ACKFrame) write(f *Framer) (err error) {
	frame.Hdr.Opcode = ACK

	if err = writeHdr(f, frame.Hdr); err != nil {
		return
	}
	if err = binary.Write(f, binary.BigEndian, frame.BlockNum); err != nil {
		return
	}
	return
}

func (frame *ERRORFrame) write(f *Framer) (err error) {
	frame.Hdr.Opcode = ERROR

	if err = writeHdr(f, frame.Hdr); err != nil {
		return
	}
	if err = binary.Write(f, binary.BigEndian, frame.ErrorCode); err != nil {
		return
	}
	if err = writeString(f, frame.ErrMsg); err != nil {
		return
	}
	return
}
