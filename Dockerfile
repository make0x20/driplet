FROM golang:1.23-alpine
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o driplet
EXPOSE 4719
CMD ["./driplet"]
