BINARY_NAME=giphy-connector

build:
	go build -o dist/$(BINARY_NAME) ./service

clean:
	rm -f $(BINARY_NAME)
