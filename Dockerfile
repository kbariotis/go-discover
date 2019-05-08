FROM scratch

COPY ./bin/discover /

EXPOSE 8080

ENTRYPOINT ["/discover"]