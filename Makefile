build:
	go build -o lens.exe \
		--ldflags="-X 'main.Version=$$(git describe --tags --exact-match || git symbolic-ref -q --short HEAD)'" \
		./server
