#background image
FROM golang:1.17-alpine

RUN apk add --no-cache go
RUN go version

#get and install packages
RUN go get -d -v ./...
RUN go install -v ./...

# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

## we run go build to compile the binary
## executable of our Go program
RUN go build -o main .

#execute main.go after creating the image
CMD ["main.go"]

