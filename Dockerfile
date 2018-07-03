FROM alpine

WORKDIR /app

COPY development.env gamelink-go ./

EXPOSE 3000

RUN apk update && apk add --no-cache ca-certificates

ENTRYPOINT [ "./gamelink-go" ]