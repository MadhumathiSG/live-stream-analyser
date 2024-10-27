package ffmpeg_analyse

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type Protocol uint

const (
	SRT Protocol = iota
	RTP
	//RTMP
)

var frameRate float32

func AnalyseInputStream(url string, protocol Protocol) error {
	fmt.Println("Start analysing the stream")
	if protocol == SRT {
		resURL := "udp://127.0.0.1:4000"
		cmd := exec.Command("ffmpeg", "-i", url, "-f", "mpegts", resURL)
		url = resURL
		if err := cmd.Start(); err != nil {
			log.Fatalf("Failed to start ffmpeg: %v", err)
		}
		//time.Sleep(2 * time.Second)
	}

	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0",
		"-show_entries", "stream=bit_rate,r_frame_rate",
		"-of", "csv=p=0", url)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to start ffprobe: %v", err)
	}
	parts := strings.Split(string(out), ",")
	fr := strings.Split(parts[0], "/")
	rate, err := strconv.Atoi(fr[0])

	fmt.Printf("Rate %v - %v\n", fr[0], rate)

	cmd = exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0",
		"-show_entries", "frame=pkt_dts_time,pkt_size,pict_type",
		"-of", "csv=p=0", url)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start ffprobe: %v", err)
	}

	reader := bufio.NewReader(stdout)
	packetSizes := []int{}
	var startTime float64
	var prevDTS float64
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		parts := strings.Split(string(line), ",")
		if len(parts) < 3 {
			continue
		}
		//
		dts, _ := strconv.ParseFloat(parts[0], 64)
		if (dts - prevDTS) > 0.034 {
			fmt.Println("Frame duration error")
		}
		size, _ := strconv.Atoi(parts[1])

		fmt.Printf("DTS: %s, Size: %s, Frame Type: %s\n",
			parts[0], parts[1], parts[2])

		if startTime == 0 {
			startTime = dts
		}

		packetSizes = append(packetSizes, size)

		// Calculate and print bitrate every second
		if dts-startTime >= 1.0 {
			totalSizeBits := 0
			for _, s := range packetSizes {
				totalSizeBits += s * 8
			}
			bitrate := float64(totalSizeBits) / (dts - startTime)
			fmt.Printf("Bitrate: %.2f kbps\n", bitrate/1000)
			packetSizes = []int{}
			startTime = dts
		}
		prevDTS = dts
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalf("ffprobe failed: %v", err)
	}
	return nil
}
