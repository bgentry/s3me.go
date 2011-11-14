package main

import (
		"flag"
		"fmt"
		"http"
		"os"
		"strconv"
)

var urlString = flag.String("u", "", "-u https://bucket.s3.amazonaws.com/key")
var segments = flag.Int("s", 0, "-s 8")
var connections = flag.Int("c", 1, "-c 4")

func main() {
		flag.Parse()
		if *urlString == "" || *segments < 0 || *connections < 1 {
				flag.Usage()
				os.Exit(1)
		}
		if *segments == 0 || *segments < *connections {
				*segments = *connections
		}
		downloadFile(*urlString, *connections, *segments)
}

func downloadFile(url string, connectionCount int, segmentCount int) {
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
		filename := "./output"
		file, err := os.OpenFile(filename, os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0600)
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
				fmt.Printf("Segment %v finished\n", i)
				finCount++
				if finCount == segmentCount {
						break
				}
		}
		fmt.Println("Download finished!")
}

type Download struct {
	url string
	file *os.File
	size int
	segments int
}

func (d *Download) SegmentSize() (int) { return d.size / d.segments }

func downloadConnection(download Download, rem chan int, fin chan int) {
		c := http.Client{}
		for n := range rem {
				offsetStart := n * download.SegmentSize()
				var offsetEnd int
				if n == (download.segments - 1) {
						offsetEnd = download.size
				} else {
						offsetEnd = ((n+1) * download.SegmentSize() - 1)
				}

				fmt.Printf("Segment %v starting\n", n)
				req, _ := http.NewRequest("GET", download.url, nil)
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offsetStart, offsetEnd))

				rsp, _ := c.Do(req)
				defer rsp.Body.Close()

				buf := make([]byte, 2048)
				writeOffset := offsetStart
				for {
// 						select {
// 						case state := <-sc:
// 								if state != StateDownloading {
// 										return
// 								}
// 						default:
// 						}

						// Read and write here...
						n_bytes, err := rsp.Body.Read(buf[0:])
						if err != nil {
								if err == os.EOF {
										break
								}
// 								return "", err  // f will be closed if we return here.
								panic(err)
						}
						nWritten, _ := download.file.WriteAt(buf[0:n_bytes], int64(writeOffset))
						writeOffset += nWritten
				}
				fin <- n
		}
}
