package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sort"
	"sync"

	"github.com/blackjack/webcam"
)

var sessionStore *SessionStore
var mutex sync.RWMutex
var jpegImage []byte

var url = flag.String("url", "", "Camera host")
var addr = flag.String("addr", ":8080", "Server address")

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Println("Start streaming")

	// Webcam begin

	cam, err := webcam.Open("/dev/video0")

	if err != nil {
		panic(err.Error())
	}
	defer cam.Close()

	format_desc := cam.GetSupportedFormats()
	var formats []webcam.PixelFormat
	for f := range format_desc {
		formats = append(formats, f)
	}

	println("Available formats: ")
	for i, value := range formats {
		fmt.Fprintf(os.Stderr, "[%d] %s\n", i+1, format_desc[value])
	}

	choice := readChoice(fmt.Sprintf("Choose format [1-%d]: ", len(formats)))
	format := formats[choice-1]

	fmt.Fprintf(os.Stderr, "Supported frame sizes for format %s\n", format_desc[format])
	frames := FrameSizes(cam.GetSupportedFrameSizes(format))
	sort.Sort(frames)

	for i, value := range frames {
		fmt.Fprintf(os.Stderr, "[%d] %s\n", i+1, value.GetString())
	}
	//choice = readChoice(fmt.Sprintf("Choose format [1-%d]: ", len(frames)))
	//size := frames[choice-1]

	f, w, h, err := cam.SetImageFormat(format, 640, 480)
	//f, w, h, err := cam.SetImageFormat(format, uint32(size.MaxWidth), uint32(size.MaxHeight))

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "Resulting image format: %s (%dx%d)\n", format_desc[f], w, h)
	}

	//println("Press Enter to start streaming")
	//fmt.Scanf("\n")

	fmt.Println("Starting streaming")

	err = cam.StartStreaming()
	if err != nil {
		panic(err.Error())
	}

	timeout := uint32(5) //5 seconds

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
				//print(".")

				//os.Stdout.Write(frame)
				//os.Stdout.Sync()
				mutex.Lock()
				jpegImage = frame
				mutex.Unlock()
			} else if err != nil {
				panic(err.Error())
			}
		}
	}()
	// webcam end

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/logout", LogoutHandler)
	http.HandleFunc("/snapshot", SnapshotHandler)
	http.HandleFunc("/live", VideoHandler)
	http.HandleFunc("/jpeg", JpegHandler)
	http.HandleFunc("/mjpeg", MotionJpegHandler)

	http.ListenAndServe(*addr, nil)
	//http.ListenAndServeTLS(*addr, "cert.pem", "key.pem", nil)
}

// Webcam related
func readChoice(s string) int {
	var i int
	for true {
		print(s)
		_, err := fmt.Scanf("%d\n", &i)
		if err != nil || i < 1 {
			println("Invalid input. Try again")
		} else {
			break
		}
	}
	return i
}

type FrameSizes []webcam.FrameSize

func (slice FrameSizes) Len() int {
	return len(slice)
}

//For sorting purposes
func (slice FrameSizes) Less(i, j int) bool {
	ls := slice[i].MaxWidth * slice[i].MaxHeight
	rs := slice[j].MaxWidth * slice[j].MaxHeight
	return ls < rs
}

//For sorting purposes
func (slice FrameSizes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func init() {
	sessionStore = NewSessionStore()
}
