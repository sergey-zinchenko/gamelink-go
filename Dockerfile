FROM alpine

WORKDIR /app

COPY gamelink-go ./

EXPOSE 3000
EXPOSE 7777

RUN apk update && apk add --no-cache ca-certificates

ENTRYPOINT [ "./gamelink-go" ]