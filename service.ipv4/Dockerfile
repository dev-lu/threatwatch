FROM golang:1.20

RUN apt update && apt upgrade -y &&\
    apt install -y git \
    make openssh-client

WORKDIR /go/src/app

COPY . ./
RUN go mod tidy \
    && go mod verify

EXPOSE 4000

RUN go build -o  /main

# Use Air for hot reload in development
RUN curl -fLo install.sh https://raw.githubusercontent.com/cosmtrek/air/master/install.sh \
    && chmod +x install.sh && sh install.sh && cp ./bin/air /bin/air
ENTRYPOINT air

#CMD ["/main"]
