package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	serverAddr := "192.168.0.6:9999" // Raspberry PiのIPアドレスとポート
	conn, err := net.Dial("udp", serverAddr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer conn.Close()

	fmt.Println("serverAddr:", serverAddr)

	var seqNum uint64 = 0
	interval := 50 * time.Millisecond

	for {
		message := fmt.Sprintf("%d", seqNum)
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Send error:", err)
			break
		}
		fmt.Println("Sent:", message)
		seqNum++
		time.Sleep(interval)
	}
}
