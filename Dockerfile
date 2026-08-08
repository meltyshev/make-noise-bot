FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /make-noise-bot .

FROM gcr.io/distroless/static-debian12

# The embedded font and wordlists carry licence terms, so their notice
# travels with the binary.
COPY --from=build /src/LICENSE /src/NOTICE /
COPY --from=build /make-noise-bot /make-noise-bot

# Config and state live in /data; mount it to keep them across restarts.
WORKDIR /data
VOLUME /data

ENTRYPOINT ["/make-noise-bot"]
