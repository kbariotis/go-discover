FROM scratch

COPY ./bin/discover /
COPY ./templates /templates

EXPOSE 8080

ENTRYPOINT ["/discover"]