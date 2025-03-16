FROM golang:1.24.0 AS builder

COPY . /src
WORKDIR /src

ENV GOPROXY=https://goproxy.cn

RUN make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/build /app
COPY --from=builder /src/configs /app/configs

WORKDIR /app

EXPOSE 8000
EXPOSE 9000

ENTRYPOINT ["./server"]
CMD ["-conf", "./configs"]
