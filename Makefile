all: build test

build:
	go build 

test:
	./discotheque 

kill:
	pkill -9 disco
	pkill -9 make
