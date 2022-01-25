#!/bin/sh

kubectl delete deployment server1-deployment
#kubectl delete deployment api2-deployment
#kubectl delete deployment api3-deployment
kubectl delete service server1-service
#kubectl delete service api2-service
#kubectl delete service api3-service
