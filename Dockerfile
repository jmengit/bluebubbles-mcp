# syntax=docker/dockerfile:1
FROM golang:1.27.3-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -trimpath -ldflags='-s -w -buildid=' -o /out/bluebubbles-mcp ./cmd/bluebubbles-mcp

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/bluebubbles-mcp /bluebubbles-mcp
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/bluebubbles-mcp"]
