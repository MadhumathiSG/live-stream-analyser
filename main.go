package main

import (
	ffmpeg_analyse "Live-Stream-Analyser/ffmpeg-analyse"
	"fmt"
)

const protocol = ffmpeg_analyse.RTP

func main() {

	var url string
	switch protocol {
	case ffmpeg_analyse.SRT:
		fmt.Println("SRT protocol")
		url = "srt://127.0.0.1:5000?mode=listener"
	case ffmpeg_analyse.RTP:
		fmt.Println("RTP protocol")
		url = "rtp://127.0.0.1:6000"
	}

	err := ffmpeg_analyse.AnalyseInputStream(url, protocol)
	if err != nil {
		fmt.Printf("Error analysing the stream: %v\n", err)
		return
	}
	//fmt.Println(res)
}
