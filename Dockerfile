# start with ubuntu based go image
FROM golang:1.21


# Install ffmpeg
RUN apt-get update && apt-get install -y ffmpeg

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go mod download


# make the build
RUN go build -o main cmd/api/main.go

RUN mkdir /app/tmp
RUN mkdir /app/tmp/converted
RUN mkdir /app/tmp/uploaded

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the executable
CMD ["./main"]

# build the image wiith this command
# docker build -t go-ffmpeg .