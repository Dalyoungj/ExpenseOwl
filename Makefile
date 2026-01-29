.PHONY: build run dev clean

build:
	go build -o expenseowl ./cmd/expenseowl

run: build
	./expenseowl -port 8080

dev:
	go run ./cmd/expenseowl -port 8080

clean:
	rm -f expenseowl
