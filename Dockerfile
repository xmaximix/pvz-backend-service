FROM golang:1.24 AS build
WORKDIR /app
COPY api/openapi.yaml api/openapi.yaml
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /pvz cmd/server/main.go

FROM gcr.io/distroless/static
WORKDIR /
COPY --from=build /pvz /
COPY --from=build /app/api/openapi.yaml /api/openapi.yaml
ENTRYPOINT ["/pvz"]