FROM alpine

WORKDIR /app

COPY gamelink-go ./

EXPOSE 3000

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

ENTRYPOINT [ "./gamelink-go" ]

