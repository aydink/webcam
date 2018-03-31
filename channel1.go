package main

import (
	"flag"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"sync"

	"github.com/blackjack/webcam"
)

var url = flag.String("url", "", "Camera host")
var addr = flag.String("addr", ":8080", "Server address")

func main() {
	flag.Parse()
	var mutex sync.RWMutex
	var jpegImage []byte

	log.Println("Start streaming")

	// Webcam begin
	cam, err := webcam.Open("/dev/video0")
	if err != nil {
		panic(err.Error())
	}
	defer cam.Close()

	// Motion JPEG format = 1196444237
	// f, w, h, err := cam.SetImageFormat(format, uint32(size.MaxWidth), uint32(size.MaxHeight))
	f, w, h, err := cam.SetImageFormat(1196444237, 1280, 720)
	fmt.Fprintf(os.Stderr, "Video format: %s (%dx%d)\n", f, w, h)

	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Started  streaming")

	err = cam.StartStreaming()
	if err != nil {
		panic(err.Error())
	}

	timeout := uint32(5) //5 seconds

	frameCounter := 0

	frameChannel := make(chan int)

	go func() {

		for {
			err = cam.WaitForFrame(timeout)

			switch err.(type) {
			case nil:
			case *webcam.Timeout:
				fmt.Fprint(os.Stderr, err.Error())
				continue
			default:
				panic(err.Error())
			}

			frame, err := cam.ReadFrame()
			if len(frame) != 0 {
				mutex.Lock()
				jpegImage = frame
				// increase frame counter
				frameCounter++
				mutex.Unlock()

				select {
				case frameChannel <- frameCounter:
				default:
				}

			} else if err != nil {
				panic(err.Error())
			}
		}
	}()
	// webcam end

	http.HandleFunc("/jpeg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(jpegImage)
	})

	http.HandleFunc("/mjpeg", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Serve streaming")

		m := multipart.NewWriter(w)
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+m.Boundary())
		w.Header().Set("Connection", "close")
		header := textproto.MIMEHeader{}

		lastFrame := 0

		for {
			if frameCounter > lastFrame {
				header.Set("Content-Type", "image/jpeg")
				header.Set("Content-Length", fmt.Sprint(len(jpegImage)))
				mw, err := m.CreatePart(header)
				if err != nil {
					break
				}
				_, err = mw.Write(jpegImage)
				if err != nil {
					break
				}

				if flusher, ok := mw.(http.Flusher); ok {
					flusher.Flush()
				}
				lastFrame = frameCounter
			} else {
				<-frameChannel
			}
		}
		log.Println("Stop streaming")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<img src="/mjpeg" />`))
	})

	http.ListenAndServe(*addr, nil)
}
