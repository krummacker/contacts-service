FROM golang:1.26.0 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build \
	-trimpath \
	-ldflags="-s -w" \
	-o /contacts-service \
	./cmd/service

FROM scratch
COPY --from=builder /contacts-service /contacts-service
ENV PORT=8080
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/contacts-service"]
