FROM golang:1.27.1-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
COPY web ./web

RUN CGO_ENABLED=0 GOOS=linux go build     -trimpath     -ldflags="-s -w -X main.version=0.0.1"     -o /out/manafield-web     ./cmd/manafield-web

FROM alpine:3.22

RUN addgroup -S -g 10001 manafield     && adduser -S -D -H -u 10001 -G manafield manafield

COPY --from=build /out/manafield-web /usr/local/bin/manafield-web

USER manafield:manafield
EXPOSE 8080

ENV PORT=8080
ENTRYPOINT ["/usr/local/bin/manafield-web"]
