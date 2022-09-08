FROM golang:alpine as builder

ENV GOPROXY https://goproxy.cn,direct

COPY ./ /source/
WORKDIR /source/

RUN go build -o powerX main.go
RUN go build -o powerX-migrate cmd/database/migrations/main.go
RUN go build -o powerX-authorization cmd/authorization/main.go cmd/authorization/openAPI.go

FROM alpine
# China mirrors
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk update --no-cache
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache tzdata
ENV TZ Asia/Shanghai
COPY --from=builder /source/powerX /app/powerX
COPY --from=builder /source/powerX-migrate /app/powerX-migrate
COPY --from=builder /source/powerX-authorization /app/powerX-authorization

RUN chmod +x /app/powerX
RUN chmod +x /app/powerX-migrate

WORKDIR /app
EXPOSE 80


ENTRYPOINT ["/app/powerX"]
