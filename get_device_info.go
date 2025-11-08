package rplidargo

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
)

func (lidar *RPLidar) GetDeviceInfo() error {
	var model, firmware_major, firmware_minor, hardware_version uint8

	err := lidar.SendRequest(0x50, nil)
	if err != nil {
		return err
	}

	data_length, multiple_response, data_type, err := ReadResponseDescriptor(lidar.SerialPort)
	if err != nil {
		return err
	}

	if multiple_response {
		return fmt.Errorf("wrong response data")
	}

	if data_type != 4 {
		return fmt.Errorf("wrong response type")
	}

	buf := make([]byte, data_length)
	_, err = lidar.SerialPort.Read(buf)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buf)

	err = binary.Read(reader, binary.LittleEndian, &model)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.LittleEndian, &firmware_minor)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.LittleEndian, &firmware_major)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.LittleEndian, &hardware_version)
	if err != nil {
		return err
	}
	strBytes := make([]byte, 16)
	if _, err := io.ReadFull(reader, strBytes); err != nil {
		return err
	}

	var reversedBytes []byte
	for i := range len(strBytes) {
		reversedBytes = append(reversedBytes, strBytes[len(strBytes)-i-1])
	}

	lidar.FirmwareMajor = int(firmware_major)
	lidar.FirmwareMinor = int(firmware_minor)
	lidar.Model = int(model)
	lidar.HardwareVersion = int(hardware_version)
	lidar.SerialNumber = hex.EncodeToString(reversedBytes)

	return nil
}
