#FROM busybox
#ADD . /usr/local/bin/my-scheduler
#WORKDIR /usr/local/bin/my-scheduler
#CMD ["./app"]

FROM python:3.5
COPY . ./athena_kube_plugin
RUN apt-get update
#RUN apt-get install python3 -y
RUN apt-get install python3-pip -v -y
RUN pip install -r ./athena_kube_plugin/requirements.txt 
#RUN python3 -m pip install numpy scipy matplotlib pandas
#RUN go get k8s.io/client-go/...

