# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.7.4

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/cloudacademy/s3zipper

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go install github.com/cloudacademy/s3zipper

COPY ./conf.json /s3zipper_conf.json

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/s3zipper -c /s3zipper_conf.json

# Document that the service listens on port 8080.
EXPOSE 7689