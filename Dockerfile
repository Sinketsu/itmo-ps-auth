# Use multi-stage build

# Create base with caching modules
FROM golang:1.13.4 AS build_base

WORKDIR /go/src/app
ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download


# Build app
FROM build_base AS builder

ENV CGO_ENABLED=0

COPY . .
RUN go build -o server


# Run app
FROM scratch

COPY --from=builder /go/src/app/server .
COPY frontend frontend
CMD ["./server"]