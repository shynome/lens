build:
	go build -o signaler.exe \
		--ldflags="-X 'main.Version=$$(git symbolic-ref -q --short HEAD || git describe --tags --exact-match)'" \
		./server
