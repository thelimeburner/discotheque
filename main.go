/*
Test program for interacting with audiomanager from go and playing audio
Can be run like so:
./discotheque -am 10.0.14.169:21403 -b 15 -media /Users/lucas/workspace/go/src/github.com/Max2Inc/SimpleAudio/media/201500.wav -z test
or to see help options
./discotheque -h
*/

package main

import (
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
var mediaPTR = flag.String("media", "/Users/lucas/workspace/go/src/github.com/Max2Inc/SimpleAudio/media/201500.wav", "Set media path")
var bufferPTR = flag.String("b", "10", "Set Buffer size in seconds")
var zonePTR = flag.String("z", "all", "Set audio zone")

const chunkSize = 1024

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

func printConfig(z Zone) {
	log.Println("Current Config:")
	log.Println("\tAudioManager: ", am.Address)
	log.Println("\tAudioManager Buffer: ", am.buffer)
	log.Println("\tZone Name: ", z.Name)
	log.Println("\tMedia File: ", media)
}

func main() {

	//parse flag commands
	flag.Parse()
	media = *mediaPTR
	//create audiomanager with addres of AM and buffer specified
	am = AudioManager{*AmPTR, nil, *bufferPTR}

	//establish a new zone
	z := Zone{*zonePTR, nil, nil, nil, nil, nil, nil}
	printConfig(z)
	//configure interrupts so we can close the program
	z.close = make(chan os.Signal, 1)
	closer := make(chan os.Signal, 1)
	signal.Notify(z.close, os.Interrupt)
	signal.Notify(closer, os.Interrupt)

	//set the buffer on the audiomanager
	am.setBuffer()

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
