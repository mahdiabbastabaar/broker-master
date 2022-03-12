#background image
#FROM golang:1.17-alpine
#FROM docker:dind


# Copy local code to the container image.
#COPY main /bin/broker

# Expecting to copy go.mod and if present go.sum.
#COPY go.* ./
#RUN go mod download

#RUN apk add --no-cache go
#RUN go version

#get and install packages
#RUN go get -d -v ./...
#RUN go install -v ./...

## we run go build to compile the binary
## executable of our Go program
#RUN go build -o main /main.go

#execute main.go after creating the image
#CMD /bin/broker

FROM golang
COPY main /bin/broker
CMD /bin/broker

