package main

import (
	"fmt"
	"log"
	"time"

	"go.bug.st/serial"
)

func print_all_received(port serial.Port) {
	for true {

		fmt.Println(ReadResponseDescriptor(port))

	}

}

func main() {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open("/dev/ttyUSB0", mode)
	if err != nil {
		log.Fatal(err)
	}
	go print_all_received(port)

	SendRequest(port, 0x50, nil)

	time.Sleep(time.Second * 10)

	port.Close()
}
