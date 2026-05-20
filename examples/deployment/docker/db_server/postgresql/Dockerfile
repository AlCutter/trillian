FROM postgres:18.4@sha256:7aba8c66107ee218ddbaacb30eae46d774bf5e34ab77a917be819b4e227493c4

# expects the build context to be: $GOPATH/src/github.com/google/trillian
COPY storage/postgresql/schema/storage.sql /docker-entrypoint-initdb.d/storage.sql
RUN chmod -R 775 /docker-entrypoint-initdb.d
