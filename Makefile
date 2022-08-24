build:
	go build -o signaler.exe \
		--ldflags="-X 'main.Version=$$(git describe --tags --exact-match || git symbolic-ref -q --short HEAD)'" \
		./server
