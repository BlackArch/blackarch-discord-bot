BIN = blackarch-bot
SOURCE = $(wildcard *.go)
LDFLAGS = -ldflags "-s -w"

build:
	@go build $(LDFLAGS) -o $(BIN) $(SRC)

clean:
	@rm -rfv $(BIN)

run: build
ifdef TOKEN
	./$(BIN) -token $(TOKEN)
else
	@echo "No token"
endif

