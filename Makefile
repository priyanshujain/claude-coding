.PHONY: build install clean

build:
	go build -o bin/claude-share ./cmd/claude-share/

install: build
	cp bin/claude-share /usr/local/bin/

clean:
	rm -rf bin/
