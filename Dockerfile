FROM golang:1.25.3

WORKDIR /bbm

COPY . .

RUN go build -o bbm-bin ./cmd/main.go

CMD ["./bbm-bin"]