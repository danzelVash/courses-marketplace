FROM golang:alpine
RUN apk add ffmpeg
WORKDIR /app
COPY . .
RUN go mod download
EXPOSE 8080
CMD ["go", "run", "cmd/main.go"]
