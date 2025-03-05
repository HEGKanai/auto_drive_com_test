package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Wi-Fi情報取得関数
func getWiFiInfo() (string, string, int, string, error) {
	ssidCmd := exec.Command("sh", "-c", "iw dev wlp1s0f0 info | grep ssid | awk '{print $2}'")
	ssidOut, err := ssidCmd.Output()
	if err != nil {
		return "", "", 0, "", err
	}
	ssid := strings.TrimSpace(string(ssidOut))

	bssidCmd := exec.Command("sh", "-c", "iw dev wlp1s0f0 link | grep 'Connected to' | awk '{print $3}'")
	bssidOut, err := bssidCmd.Output()
	if err != nil {
		return "", "", 0, "", err
	}
	bssid := strings.TrimSpace(string(bssidOut))

	rssiCmd := exec.Command("sh", "-c", "iw dev wlp1s0f0 link | grep 'signal' | awk '{print $2}'")
	rssiOut, err := rssiCmd.Output()
	if err != nil {
		return "", "", 0, "", err
	}
	rssiStr := strings.TrimSpace(string(rssiOut))
	rssi, err := strconv.Atoi(rssiStr)
	if err != nil {
		return "", "", 0, "", err
	}

	frqCmd := exec.Command("sh", "-c", "iw dev wlp1s0f0 info | grep channel")
	frqOut, err := frqCmd.Output()
	if err != nil {
		return "", "", 0, "", err
	}
	frq := strings.TrimSpace(string(frqOut))

	return ssid, bssid, rssi, frq, nil
}

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
	writer := bufio.NewWriter(logFile)

	var prevSSID, prevBSSID, prevFRQ string
	var prevRSSI int

	fmt.Println("Process Started...")

	for {
		// UDPパケット受信
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
			writer.WriteString(logMsg)
		}

		// 正常に受信したパケットをログに記録
		writer.WriteString(fmt.Sprintf("Received packet %d at %s\n", seqNum, time.Now().Format(time.RFC3339)))
		lastSeqNum = seqNum

		// Wi-Fi状態の取得
		ssid, bssid, rssi, frq, err := getWiFiInfo()
		if err != nil {
			fmt.Println("Failed to get Wi-Fi info:", err)
			continue
		}

		// AP変更や電波強度変化の記録
		if ssid != prevSSID || bssid != prevBSSID || rssi != prevRSSI || frq != prevFRQ {
			logMsg := fmt.Sprintf("Wi-Fi status changed: SSID=%s, BSSID=%s, Signal Strength=%d dBm, Channel = %s at %s\n",
				ssid, bssid, rssi, frq, time.Now().Format(time.RFC3339))
			fmt.Print(logMsg)
			writer.WriteString(logMsg)

			// 更新
			prevSSID = ssid
			prevBSSID = bssid
			prevRSSI = rssi
			prevFRQ = frq
		}

		writer.Flush()
	}
}
