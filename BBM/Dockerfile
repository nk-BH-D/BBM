FROM golang:1.24.3

WORKDIR /bbm

COPY . .

RUN go build -o bbm-bin ./cmd/main.go

CMD ["./bbm-bin"]