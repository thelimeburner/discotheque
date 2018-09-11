/*
Test program for interacting with audiomanager from go and playing audio
Can be run like so:
./discotheque -am 10.0.14.169:21403 -b 15 -media /Users/lucas/workspace/go/src/github.com/Max2Inc/SimpleAudio/media/201500.wav -z test
or to see help options
./discotheque -h
*/

package main

import (
	"bufio"
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/sacOO7/gowebsocket"
)

var media string

var am AudioManager //{"10.0.14.169:21403", nil, "20"} //set buffer size

//flag for address to serve content
var AmPTR = flag.String("am", "10.0.14.169:21403", "Set audioManager address/port")
var mediaPTR = flag.String("media", "/media/song.wav", "Set media path")
var bufferPTR = flag.String("b", "4", "Set Buffer size in seconds")
var zonePTR = flag.String("z", "all", "Set audio zone")
var filePTR = flag.String("f", "", "Set file to play from")

// const chunkSize = 35280
const chunkSize = 44100 / 8 / 2 //(8 * 44100 * 2 * 0.05) / 8 // 1024 * 3
// const chunkSize = 1024
const chunkSizeForHalf = (16 * 44100 * 2 * 0.1) / 8

// const chunkSize = (16 * 44100 * 2 * 2) / 8

type AudioManager struct {
	Address string
	i       chan os.Signal
	buffer  string
}

type Zone struct {
	Name  string
	clip  *os.File
	start chan bool
	stop  chan bool
	close chan os.Signal
	ws    *gowebsocket.Socket
	bytes []byte
}

var stopped chan bool

func printConfig(z Zone) {
	log.Println("Current Config:")
	log.Println("\tAudioManager: ", am.Address)
	log.Println("\tAudioManager Buffer: ", am.buffer)
	log.Println("\tZone Name: ", z.Name)
	log.Println("\tMedia File: ", media)
}

func readClips(fname string) []string {
	//open file for reading
	file, err := os.Open(fname)
	if err != nil {
		log.Println("Error opening file: ", err.Error())
		os.Exit(1)
	}
	defer file.Close()

	var lines []string

	//open scanner for and read into lines
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

//play clips from a list in a file
func (z *Zone) playFileList(fname string) {
	clips := readClips(fname)
	//initialize the zone and its websocket options
	z.init()
	for _, cl := range clips {
		log.Println("Now playing clip: " + cl)
		//open the clip and return the file pointer
		z.clip = openClip(cl)
		//start player, note doesnt play until connected and receive start message
		go z.play()
		z.start <- z.reconnect()
		//establish websocket connections
		z.reconnect()
		<-stopped
	}

}

func (z *Zone) playSingleSong(fname string) {
	//open the clip and return the file pointer
	z.clip = openClip(media)

	//initialize the zone and its websocket options
	z.init()

	//start player, note doesnt play until connected and receive start message
	go z.play()

	//establish websocket connections
	z.reconnect()

	//prevent exit
	for {
		select {
		case <-closer:
			os.Exit(0)
		default:
			continue
		}

	}
}

var closer chan os.Signal

func main() {

	//parse flag commands
	flag.Parse()
	media = *mediaPTR

	//create audiomanager with addres of AM and buffer specified
	am = AudioManager{*AmPTR, nil, *bufferPTR}

	log.Println(chunkSize)

	//establish a new zone
	z := Zone{*zonePTR, nil, nil, nil, nil, nil, nil}
	printConfig(z)
	//configure interrupts so we can close the program
	z.close = make(chan os.Signal, 1)
	closer = make(chan os.Signal, 1)
	stopped = make(chan bool, 1)
	signal.Notify(z.close, os.Interrupt)
	signal.Notify(closer, os.Interrupt)

	//set the buffer on the audiomanager
	am.setBuffer()
	if *filePTR == "" {
		log.Println("Playing single clip: " + media)
		z.playSingleSong(media)
	} else {
		log.Println("Playing list of clips from : " + *filePTR)
		z.playFileList(*filePTR)
	}

}

//open a clip using general file opener
func openClip(media string) *os.File {

	f2, err := os.Open(media)
	// Check for errors when opening the file
	if err != nil {
		log.Fatal(err)
	}
	//defer f2.Close()

	return f2
}

//makes a request to
func (am *AudioManager) setBuffer() bool {

	request := "http://" + am.Address + "/audio/buffer/capacity/" + am.buffer
	log.Println("Making Buffer Request", request)
	msg := []byte("")
	resp, err := http.Post(request, "body/type", bytes.NewBuffer(msg))
	if err != nil {
		log.Println("setBuffer Request Failed")
		log.Println("Using Default buffer")
		log.Println(err)
		return false

	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		log.Println("setBuffer Read Failed")
		log.Println("Using Default buffer")
		log.Println(err)
		return false
	}
	log.Println(string(body))
	bodyStr := string(body)
	//var d []interface{}
	if bodyStr == "Resource not found" {
		log.Println("setBuffer Failed")
		log.Println("Using Default buffer")
		return false
	}
	if bodyStr == "Buffer capacity set" {

		log.Println("Buffer set to ", am.buffer)
		return true
	}

	return false
}
