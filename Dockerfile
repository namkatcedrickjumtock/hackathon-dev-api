FROM ubuntu:20.04


RUN apt-get update && apt-get install -y ca-certificates openssl

# Set destination for COPY
WORKDIR /app/bin

# specif location to copy go binary
COPY ./bin/api ./
COPY ./db ./
ENV DB_MIGRATIONS_PATH=./db/migrations

ENV LISTEN_PORT=9080
EXPOSE  9980

# Run
CMD ["/app/bin/api" ]
