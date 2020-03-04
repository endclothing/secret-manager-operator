FROM golang:1.13-alpine as builder
RUN adduser --uid 1993 --disabled-password scratchuser
RUN apk add git
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main main.go

FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/ /app/
WORKDIR /app
USER 1993
CMD ["./main"]
