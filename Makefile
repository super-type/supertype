build:
	go build -o bin/main cmd/supertype/main.go

run:
	cat supertype.txt
	go run cmd/supertype/main.go