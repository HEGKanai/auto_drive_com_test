package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

func main() {
	port := ":9999"
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	var lastSeqNum uint64 = 0
	logFile, err := os.Create("packet_loss_log.txt")
	if err != nil {
		fmt.Println("Failed to create log file:", err)
		return
	}
	defer logFile.Close()

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Receive error:", err)
			continue
		}

		seqStr := string(buffer[:n])
		seqNum, err := strconv.ParseUint(seqStr, 10, 64)
		if err != nil {
			fmt.Println("Invalid sequence number:", seqStr)
			continue
		}

		// 欠落パケットの検出
		if lastSeqNum != 0 && seqNum > lastSeqNum+1 {
			lossCount := seqNum - lastSeqNum - 1
			logMsg := fmt.Sprintf("Packet loss detected: lost %d packets between %d and %d at %s\n",
				lossCount, lastSeqNum, seqNum, time.Now().Format(time.RFC3339))
			fmt.Print(logMsg)
			logFile.WriteString(logMsg)
		}

		// 正常に受信したパケットをログに記録
		logFile.WriteString(fmt.Sprintf("Received packet %d at %s\n", seqNum, time.Now().Format(time.RFC3339)))
		lastSeqNum = seqNum
	}
}
