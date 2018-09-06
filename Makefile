all: clean build test

2: clean build test2

build:
	go build 

test:
	./discotheque 

test2:
	./discotheque -media /Users/lucas/workspace/go/src/github.com/Max2Inc/SimpleAudio/media/stereo.wav

clean: 
	rm -rf discotheque
kill:
	pkill -9 disco
	pkill -9 make
