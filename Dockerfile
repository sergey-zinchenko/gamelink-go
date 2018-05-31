FROM scratch

ENV PORT 3000
EXPOSE $PORT

COPY gamelink-go /
CMD ["/gamelink-go"]