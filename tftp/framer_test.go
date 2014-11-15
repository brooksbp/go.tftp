package tftp

import (
	"bytes"
	"reflect"
	"testing"
)

func TestCreateParseRRQ(t *testing.T) {
	buffer := new(bytes.Buffer)
	framer := NewFramer(buffer)

	rrqFrame := RRQFrame{
		Hdr: FrameHdr{
			Opcode: RRQ,
		},
		Filename: "test.png",
		Mode:     "octet",
	}
	if err := framer.WriteFrame(&rrqFrame); err != nil {
		t.Fatal("WriteFrame:", err)
	}

	frame, err := framer.ReadFrame()
	if err != nil {
		t.Fatal("ReadFrame:", err)
	}
	parsedRrqFrame, ok := frame.(*RRQFrame)
	if !ok {
		t.Fatal("Parsed incorrect frame type:", frame)
	}
	if !reflect.DeepEqual(rrqFrame, *parsedRrqFrame) {
		t.Fatal("got:", *parsedRrqFrame, "\nwant:", rrqFrame)
	}
}
