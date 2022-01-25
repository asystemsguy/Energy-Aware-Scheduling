kill $(lsof -t -i:8090)
kubectl port-forward svc/server-a-service -n athena 8090:8090 &
sleep 10
ab -n 10 -c 10 -s 5000 -m GET -g bench http://localhost:8090/a
