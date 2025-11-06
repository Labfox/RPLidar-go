package rplidargo

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func (lidar *RPLidar) GetHealth() (int, int, error) {
	if lidar.IsMock {
		return 0,0, nil
	}
	err := lidar.SendRequest(0x52, nil)
	if err != nil {
		return 0, 0, err
	}

	data_length, multiple_response, data_type, err := ReadResponseDescriptor(lidar.SerialPort)
	if err != nil {
		return 0, 0, err
	}

	if multiple_response {
		return 0, 0, fmt.Errorf("Wrong response data")
	}

	if data_type != 6 {
		return 0, 0, fmt.Errorf("Wrong response type")
	}

	buf := make([]byte, data_length)
	_, err = lidar.SerialPort.Read(buf)
	if err != nil {
		return 0, 0, err
	}

	reader := bytes.NewReader(buf)

	var status uint8
	err = binary.Read(reader, binary.LittleEndian, &status)
	if err != nil {
		return 0, 0, err
	}

	var error_code uint16
	err = binary.Read(reader, binary.LittleEndian, &error_code)
	if err != nil {
		return int(status), 0, err
	}

	return int(status), int(error_code), nil
}
