all: clean build test

2: clean build test2
3: clean build test3

build:
	go build 

test:
	./discotheque 

test2:
	./discotheque -media /Users/lucas/workspace/go/src/github.com/Max2Inc/SimpleAudio/media/stereo.wav -b 20

test3:
	./discotheque -media charlieBoy.wav -b 20


single: clean build
	./discotheque -media clips/killed.wav -am am:21403 -b 4

list: clean build
	./discotheque -f cliplist.txt -am am:21403 -b 4

clean: 
	rm -rf discotheque
kill:
	pkill -9 disco
	pkill -9 make
