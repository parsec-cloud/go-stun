FROM golang:1.17.7-buster
ENV STUN_USER STUN
ENV STUN_HOME /go/src/stun
ARG GROUP_ID
ARG USER_ID
RUN groupadd --gid $GROUP_ID STUN && useradd -m -l --uid $USER_ID --gid $GROUP_ID $STUN_USER
RUN mkdir -p $STUN_HOME && chown -R $STUN_USER:$STUN_USER $STUN_HOME
USER $STUN_USER
WORKDIR $STUN_HOME
RUN mkdir -p stun-server health-server
COPY ./stun-server/main.go stun-server/main.go
COPY ./health-server/main.go health-server/main.go
COPY run_services.sh run_services.sh
EXPOSE 3478/udp
EXPOSE 8888/tcp
RUN GO111MODULE=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o stun-server/service stun-server/main.go 
RUN GO111MODULE=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o health-server/service health-server/main.go 
CMD ["./run_services.sh"]

# docker build --build-arg USER_ID=1234 --build-arg GROUP_ID=1234 -t stun-server .
# docker run -it --rm -p 8888:8888/tcp -p 3478:3478/udp stun-server
