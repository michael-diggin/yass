FROM golang:1.14.6-alpine as builder

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
EXPOSE 8080
RUN GOOS=linux go build -o /bin/server github.com/michael-diggin/yass/backend/cmd

ENTRYPOINT [ "/bin/server" ]