FROM golang
WORKDIR /src/service
COPY . .
RUN apt-get update && apt-get install fuse libfuse-dev -y
RUN CGO_ENABLED=1 GOOS=linux go build  \
    -ldflags "-X main.Version=`git tag --sort=-version:refname | head -n 1`"  \
    -o rms-music-bot -a -installsuffix cgo rms-music-bot.go \
    && mkdir /app && cp rms-music-bot /app/  \
    && mkdir -p /etc/rms && cp configs/rms-music-bot.json /etc/rms/

WORKDIR /app
CMD ["./rms-music-bot"]
