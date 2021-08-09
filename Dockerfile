FROM golang:1.16-alpine as builder
WORKDIR /go/src/app
COPY . ./
RUN CGO_ENABLED=0 go build -a -ldflags '-s -w -extldflags "-static"' -o bin/app-server .

FROM scratch
COPY --from=builder /go/src/app/bin/app-server /app-server
COPY --from=builder /etc/passwd /etc/passwd
USER nobody
CMD ["/app-server"]