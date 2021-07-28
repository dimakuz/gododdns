FROM golang:latest AS builder

ADD . /code
WORKDIR /code
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"'

FROM scratch
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=builder /code/gododdns /gododdns
ENTRYPOINT [ "/gododdns" ]