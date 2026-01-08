FROM golang:1.25 AS builder

WORKDIR /src
COPY .. .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /example ./examples


FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /example .
USER 65532:65532

ENTRYPOINT ["/example"]
