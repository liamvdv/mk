build:
	go build -o mk
install: build
	mv ./mk ~/.local/bin
env:
	export _MK_DIR_EDITOR="code"
	export _MK_FILE_EDITOR="code"
