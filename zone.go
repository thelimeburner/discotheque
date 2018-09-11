package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sacOO7/gowebsocket"
)

//reconnect checks if we are connected still through the ws. if so return true other wise attempt to reconnect 15 times each time sleeping for 2 seconds
func (z *Zone) reconnect() bool {
	if z.ws.IsConnected {
		return true
	}
	counter := 0
	for !z.ws.IsConnected {
		if counter > 15 {
			return false
		}
		time.Sleep(2 * time.Second)
		log.Println("[INFO] Attempting to reconnect")
		z.ws.Connect()
		counter++

	}

	return true

}

//sets up websocket behavior. DOES NOT ACTUALLY CONNECT
func (z *Zone) init() {

	//define websocket
	socket := gowebsocket.New("ws://" + am.Address + "/audio/" + z.Name + "/socket")
	//store websockets
	z.ws = &socket

	//make channels to start and stop audio from AM
	z.start = make(chan bool)
	z.stop = make(chan bool)

	//set socket options. These are mandatory! Otherwise bad handshake error
	z.ws.RequestHeader.Set("Accept-Encoding", "gzip, deflate")
	z.ws.RequestHeader.Set("Accept-Language", "en-US,en;q=0.9")
	z.ws.RequestHeader.Set("Origin", "")
	z.ws.ConnectionOptions = gowebsocket.ConnectionOptions{
		UseSSL:         false,
		UseCompression: true,
		Subprotocols:   []string{"grut"}, //this is necessary!!
	}

	/* These control message handling from websockets*/
	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")
		z.start <- true
	}

	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Println("Recieved connect error ", err)
	}

	//handle receive of messages when we get start tell player, when we receieve stop tell player
	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		log.Println("Recieved message " + message)
		switch message {
		case "start":
			//z.sendData = true
			z.start <- true

			//go z.Stream()
		case "stop":
			z.stop <- false
			//z.sendData = false
			// stop <- true
			// <-stopped

		}

	}

	socket.OnBinaryMessage = func(data []byte, socket gowebsocket.Socket) {
		log.Println("Recieved binary data ", data)
	}

	socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Recieved ping " + data)
	}

	socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Recieved pong " + data)
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
		return
	}
}

//this is the actual player functionality
func (z *Zone) play() {

	status := false
	b := make([]byte, chunkSize) // make a buffer to handle read
	for {
		select {
		case <-z.stop: //receive stop message, set status to false
			fmt.Println("Received Stop Message!")
			status = false
		case <-z.start: //receive start message, set status to start
			fmt.Println("Received start Message!")
			status = true
		case <-z.close: //receieve close from interrupt so exit!
			os.Exit(0)
			return
		default: //received no messages continue as normal

			if status { // if true we are able to stream.

				n, err := z.clip.Read(b) // read into buffer
				if err != nil {          // if failed to readin we are at end of file so exit.
					log.Println(err.Error())
					log.Println("End of file!")
					os.Exit(0)
				}

				z.ws.SendBinary(b)                                                                                               // send bytes we read in.
				log.Printf("Sent bytes: %d. First byte: %d Middle Byte: %d Last Byte: %d", n, b[0], b[chunkSize/2], b[len(b)-2]) // print that we sent.
				time.Sleep(5 * time.Millisecond)
			}
		}
	}

}
