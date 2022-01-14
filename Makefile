BINARY_NAME=giphy-connector

build:
	go build -o dist/$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
