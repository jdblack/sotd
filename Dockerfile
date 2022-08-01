FROM golang:1.18

WORKDIR /usr/src/sotd

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/sotd ./...

CMD ["sotd"]

