FROM golang:1.14.6

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
EXPOSE 8080
RUN GOOS=linux go build -o server github.com/michael-diggin/yass/backend/cmd/main.go

ENTRYPOINT [ "./server" ]