# start with ubuntu based go image
FROM golang:1.21

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go mod download

#add ffmpeg
RUN apk add --no-cache ffmpeg

# install make
RUN apk add --no-cache make

# make the build
RUN go build -o main cmd/api/main.go

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the executable
CMD ["main"]

# build the image wiith this command
# docker build -t go-ffmpeg .