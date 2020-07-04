BIN = blackarch-bot
SOURCE = $(wildcard *.go)
LDFLAGS = -ldflags "-s -w"

default:
	@echo 'check the Makefile for clues'

clean:
	@rm -rfv $(BIN)

build:
	@go build $(LDFLAGS) -o $(BIN) $(SRC)

