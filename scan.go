package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// TODO: Fix error handling
func (lidar *RPLidar) ReadScans(readLength int) error {
	fmt.Println("launch reading")
	for true {
		buf := make([]byte, readLength)
		_, err := lidar.SerialPort.Read(buf)
		if err != nil {

			return err
		}

		reader := bytes.NewReader(buf)

		// Extract first byte

		var b [1]byte
		_, err = reader.Read(b[:1])
		if err != nil {
			return err
		}
		combined := b[:][0]

		quality := combined >> 2

		flagA := (combined >> 1) & 0x1
		flagB := (combined >> 0) & 0x1

		if flagA == flagB {
			continue
		}

		newScan := flagA == 0

		// Extract second byte

		var b2 [2]byte
		_, err = reader.Read(b2[:2])
		if err != nil {
			return err
		}
		combined2 := binary.LittleEndian.Uint16(b2[:])

		angle_f1 := (combined2 >> 1) / 64

		flagC := (combined2 >> 0) & 0x1

		if flagC != 1 {
			continue
		}

		// Read last field

		var b3 [2]byte
		_, err = reader.Read(b3[:2])
		if err != nil {
			return err
		}
		distance := binary.LittleEndian.Uint16(b3[:])

		distance = distance / 4

		lidar.DistanceReadings <- DistanceReading{
			Quality:  int(quality),
			NewScan:  newScan,
			Angle:    float32(angle_f1),
			Distance: float32(distance),
		}
	}

	return nil
}

func (lidar *RPLidar) Scan() error {
	err := lidar.SerialPort.SetDTR(false)
	if err != nil {
		return err
	}

	err = lidar.SendRequest(0x20, nil)
	if err != nil {
		return err
	}

	data_length, multiple_response, data_type, err := ReadResponseDescriptor(lidar.SerialPort)
	if err != nil {
		return err
	}

	if !multiple_response {
		return fmt.Errorf("Wrong response data")
	}

	if data_type != 0x81 {
		return fmt.Errorf("Wrong response type")
	}

	go lidar.ReadScans(int(data_length))

	return nil
}
