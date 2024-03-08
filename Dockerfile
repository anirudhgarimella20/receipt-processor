FROM golang:1.19

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go .
EXPOSE 8080

RUN go build -o bin 

ENTRYPOINT ["/app/bin"]

