# syntax=docker/dockerfile:1
FROM golang:1.20-alpine
RUN apk --no-cache add tzdata

WORKDIR /src/

COPY go.mod go.sum /src/

RUN go mod download

COPY pkg /src/pkg

ARG MAINFILE=butlerbot
COPY cmd/${MAINFILE} /src/cmd/${MAINFILE}
RUN cd /src/cmd/${MAINFILE} && CGO_ENABLED=0 GOOS=linux go build -o /src/main .

FROM scratch

COPY --from=0 /src/main /

WORKDIR /
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /etc/passwd /etc/passwd
COPY --from=0 /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=America/Vancouver
USER nobody

ENTRYPOINT [ "/main" ]
