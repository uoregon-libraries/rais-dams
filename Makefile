.PHONY: dams

dams:
	go build -ldflags="-s -w" -o ./bin/dams
