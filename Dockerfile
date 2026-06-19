# syntax=docker/dockerfile:1
FROM golang:1.26.4-alpine3.22 AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILDDATE=unknown
ARG BRANCH=unknown

RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X github.com/hrodrig/kiko/internal/version.Version=${VERSION} -X github.com/hrodrig/kiko/internal/version.Commit=${COMMIT} -X github.com/hrodrig/kiko/internal/version.BuildDate=${BUILDDATE} -X github.com/hrodrig/kiko/internal/version.Branch=${BRANCH}" -o /usr/local/bin/kiko ./cmd/kiko

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /usr/local/bin/kiko /usr/local/bin/kiko

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/kiko"]
CMD ["serve"]
