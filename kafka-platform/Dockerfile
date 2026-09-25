# ---------- build ----------
FROM golang:1.27-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates git

# Copy everything including vendor/
COPY . .

# Build from vendored modules — no network, no proxy, no surprises
RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /out/producer ./cmd/producer

# ---------- runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=build /out/producer /app/producer
COPY --from=build /src/schemas /app/schemas

USER nonroot:nonroot
EXPOSE 9090
ENTRYPOINT ["/app/producer"]