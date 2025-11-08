package rplidargo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"go.bug.st/serial"
)

type RPLidar struct {
	Model            int
	FirmwareMajor    int
	FirmwareMinor    int
	HardwareVersion  int
	SerialNumber     string
	SerialPort       serial.Port
	DistanceReadings chan DistanceReading
	IsMock           bool
}

type DistanceReading struct {
	Quality  int
	NewScan  bool
	Angle    float32
	Distance float32
}

func NewRPLidar(device string, baudRate int) (RPLidar, error) {
	var nRPLidar RPLidar = RPLidar{
		DistanceReadings: make(chan DistanceReading),
	}

	var err error

	if device == "mock" {
		nRPLidar.IsMock = true

	} else {
		mode := &serial.Mode{
			BaudRate: baudRate,
		}
		nRPLidar.SerialPort, err = serial.Open(device, mode)
		if err != nil {
			return nRPLidar, err
		}

		err = nRPLidar.GetDeviceInfo()
		if err != nil {
			return nRPLidar, err
		}

	}

	return nRPLidar, nil
}

func (lidar *RPLidar) SendRequest(command byte, data *[]byte) error {
	buf := new(bytes.Buffer)

	buf.Write([]byte{0xA5, command})

	if data == nil {
		_, err := lidar.SerialPort.Write(buf.Bytes())
		return err
	}

	dSize := uint8(len(*data))
	binary.Write(buf, binary.LittleEndian, dSize)

	buf.Write(*data)

	checksum := byte(0)
	checksum ^= 0xA5
	checksum ^= command
	checksum ^= byte(len(*data))

	for _, b := range *data {
		checksum ^= b
	}

	buf.Write([]byte{checksum})

	_, err := lidar.SerialPort.Write(buf.Bytes())
	return err
}

// TODO: Fix the horrible bits reading
func ReadResponseDescriptor(serial serial.Port) (data_length uint32, multiple_response bool, data_type byte, err error) {
	buf := make([]byte, 7)
	serial.Read(buf)

	reader := bytes.NewReader(buf)

	startflag1 := make([]byte, 1)

	binary.Read(reader, binary.LittleEndian, &startflag1)

	if startflag1[0] != 0xA5 {

		return 0, false, 0x00, fmt.Errorf("invalid response flag (1)")
	}

	startflag2 := make([]byte, 1)

	binary.Read(reader, binary.LittleEndian, &startflag2)

	if startflag2[0] != 0x5A {

		return 0, false, 0x00, fmt.Errorf("invalid response flag (2)")
	}

	var b [4]byte
	_, err = reader.Read(b[:4])
	if err != nil {
		log.Fatal(err)
	}
	combined := binary.LittleEndian.Uint32(b[:])

	length := combined & 0x3FFFFFFF        // lower 30 bits
	send_mode_bt := (combined >> 30) & 0x3 // upper 2 bits

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
