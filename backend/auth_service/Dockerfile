FROM golang:1.24-alpine

WORKDIR /auth_service

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o auth ./app/cmd

EXPOSE 8080

CMD ["./auth"]