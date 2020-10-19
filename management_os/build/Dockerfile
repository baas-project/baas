FROM golang as build

WORKDIR /build

# Force modules
ENV GO111MODULE=on

# Cache dependencies
COPY programs/go.* ./
RUN go mod download

# Build project
COPY programs .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o programs

# ---------------------

FROM debian:buster-slim
COPY --from=build /build/programs programs
