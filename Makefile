build:
	go build -o mk
install: build
	mv ./mk ~/.local/bin 