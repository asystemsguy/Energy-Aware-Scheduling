#FROM busybox
#ADD . /usr/local/bin/my-scheduler
#WORKDIR /usr/local/bin/my-scheduler
#CMD ["./app"]

FROM debian
RUN apt-get update
RUN apt-get install curl
COPY ./app /app
ENTRYPOINT /app