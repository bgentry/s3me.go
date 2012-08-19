package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

var urlString = flag.String("u", "", "-u https://bucket.s3.amazonaws.com/key")
var outputPath = flag.String("o", "./output", "Output path")
var segments = flag.Int("s", 0, "Number of segments")
var connections = flag.Int("c", 1, "Number of connections")

func main() {
	flag.Parse()
	if *urlString == "" || *segments < 0 || *connections < 1 {
		flag.Usage()
		os.Exit(1)
	}
	if *segments == 0 || *segments < *connections {
		*segments = *connections
	}
	downloadFile(*urlString, *outputPath, *connections, *segments)
}

func downloadFile(url string, filename string, connectionCount int, segmentCount int) {
	fmt.Printf("Fetching HEAD from %p\n", url)
	resp, err := http.Head(url)
	if err != nil {
		fmt.Printf("ERROR: %p\n", err)
		os.Exit(1)
	}
	fmt.Println(resp)
	fmt.Println(resp.Header["Content-Length"][0])
	fileSize, err := strconv.Atoi(resp.Header["Content-Length"][0])
	if err != nil {
		fmt.Printf("Invalid Content-Length: %p\n", err)
		os.Exit(1)
	}

	remainingChan := make(chan int, segmentCount)
	finChan := make(chan int, segmentCount)
	for i := 0; i < segmentCount; i++ {
		remainingChan <- i
	}
	close(remainingChan)

	// Open file for writing
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(fmt.Sprintf("Error opening %v\n", filename))
	}
	defer file.Close()

	download := Download{url, file, fileSize, segmentCount}
	// Initiate downloads
	for i := 0; i < connectionCount; i++ {
		go downloadConnection(download, remainingChan, finChan)
	}

	finCount := 0
	for i := range finChan {
		fmt.Printf("Segment %v DONE %v\n", i, time.UTC().Format(time.RFC3339))
		finCount++
		if finCount == segmentCount {
			break
		}
	}
	fmt.Printf("Download finished! %v\n", time.UTC().Format(time.RFC3339))
}

type Download struct {
	url      string
	file     *os.File
	size     int
	segments int
}

func (d *Download) SegmentSize() int { return d.size / d.segments }

func downloadConnection(download Download, rem chan int, fin chan int) {
	c := http.Client{}
	for n := range rem {
		offsetStart := n * download.SegmentSize()
		var offsetEnd int
		if n == (download.segments - 1) {
			offsetEnd = download.size
		} else {
			offsetEnd = ((n+1)*download.SegmentSize() - 1)
		}

		fmt.Printf("Segment %v BEGIN %v\n", n, time.UTC().Format(time.RFC3339))
		req, _ := http.NewRequest("GET", download.url, nil)
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offsetStart, offsetEnd))

		rsp, _ := c.Do(req)
		defer rsp.Body.Close()

		buf := make([]byte, 2048)
		writeOffset := offsetStart
		for {
			n_bytes, err := rsp.Body.Read(buf[0:])
			if err != nil {
				if err == os.EOF {
					break
				}
				panic(err)
			}
			nWritten, _ := download.file.WriteAt(buf[0:n_bytes], int64(writeOffset))
			writeOffset += nWritten
		}
		fin <- n
	}
}
