package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/icza/bitio"
	"go.bug.st/serial"
)

func SendRequest(serial serial.Port, command byte, data *[]byte) error {
	buf := new(bytes.Buffer)

	buf.Write([]byte{0xA5, command})

	if data == nil {
		serial.Write(buf.Bytes())
		return nil
	} else {
		panic(fmt.Errorf("not implemented yet"))
	}
}

func ReadResponseDescriptor(serial serial.Port) (data_length int32, multiple_response bool, data_type byte, err error) {
	buf := make([]byte, 7)
	serial.Read(buf)

	reader := bytes.NewReader(buf)

	startflag1 := make([]byte, 1)

	binary.Read(reader, binary.LittleEndian, &startflag1)

	if startflag1[0] != 0xA5 {
		return 0, false, 0x00, fmt.Errorf("Invalid response flag (1)")
	}

	startflag2 := make([]byte, 1)

	binary.Read(reader, binary.LittleEndian, &startflag2)

	if startflag2[0] != 0x5A {
		return 0, false, 0x00, fmt.Errorf("Invalid response flag (2)")
	}

	bitsReader := bitio.NewReader(reader)

	length_sb, err := bitsReader.ReadBits(30)
	if err != nil {
		return 0, false, 0x00, err
	}

	length_byte := int32(length_sb)

	bb_writer := new(bytes.Buffer)

	binary.Write(bb_writer, binary.BigEndian, length_byte)

	breader := bytes.NewReader(bb_writer.Bytes())

	var length int32

	binary.Read(breader, binary.LittleEndian, &length)

	fmt.Println(length_sb, length_byte, bb_writer.Bytes(), length)
	send_mode_bt, err := bitsReader.ReadBits(2)
	if err != nil {
		return length, false, 0x00, err
	}

	switch send_mode_bt {
	case 0x0:
		multiple_response = false
	case 0x1:
		multiple_response = true
	}

	var dtype byte

	binary.Read(reader, binary.LittleEndian, &dtype)

	return length, multiple_response, dtype, nil
}
