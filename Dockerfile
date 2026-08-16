# Heddle builds offline and ships as a single static binary.
#
# The builder stage compiles with the module proxy turned off and no network access, so the
# image content is a function of the source tree alone. The final stage is scratch: the tool
# reads a draft from its flags or from a file and writes grids of characters, and needs no
# shell, no certificates and no user database.
FROM golang:1.22 AS builder

ENV GOTOOLCHAIN=local \
    CGO_ENABLED=0 \
    GOPROXY=off \
    GOFLAGS=-mod=mod

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN go vet ./... \
 && go test ./... \
 && go build -trimpath -ldflags "-s -w" -o /out/heddle ./cmd/heddle

FROM scratch

COPY --from=builder /out/heddle /heddle
COPY examples /examples

ENTRYPOINT ["/heddle"]
CMD ["--help"]
