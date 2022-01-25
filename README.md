### Prerequisites
  You need Go version **v1.10** or higher.  
  To build, make sure Go dependencies are installed.    
  To get dependencies, run  ```go get -d ./...```  

  if it does not work, install package separately  
  (kuberenetes client-go)  
  ```go get k8s.io/client-go/...```  
  ```go get -u k8s.io/apimachinery/...```
  
  (prometheus client-go)  
  ```go get  github.com/prometheus/..```

  Verify for below folders of Kubernetes client-go library and prometheus after running above commands  
  **$GOPATH**/src/k8s.io/client-go/  
  **$GOPATH**/src/github.com/prometheus/
  
  Update URLs for the Prometheus server in the file **kubernetes-client/scheduler.go** in the variables: **promeMonitorAdr** and **promeIstioAdr**

### Add tags to machines in the Kubernetes Cluster
Custom scheduling algorithm reads the tags of the machines in the cluster to make scheduling decisions. To add a tag to machine, run ```kubectl label nodes ${node_name} VM=${VMn} PM=${PMn} RM=${RMn}```.


### Build the Scheduler Code
  ```cd kubernetes-client```  
  ```make build```  
  
### Deploy Benchmark application
  1. Run ```kubectl create namespace athena```
  2. To deploy using default Kubernetes scheduler, run
    ```make rund```
  3. To deploy using custom scheduler, run
    ```make runa```
  
### Deploy Robot shop application
  1. Run```kubectl create namespace athena```
  2. To deploy using default Kubernetes scheduler, run
    ```make runrobod```
  3. To deploy using custom scheduler, run
    ```make runroboa```


### Test the deployed application
To test the application from outside the Kubernetes cluster, we need to do port forwarding.

For benchmark application:    
Port forwarding - ```kubectl port-forward svc/server-a-service ${port A}:8090 -n athena```    
To test,send GET http request to http://{url}:${port A}/a

For Robot Shop applcation:  
Port forwarding - ```kubectl port-forward svc/web ${port B}:8080 -n athena```   
To test, send GET http request to http://{url}:${port B}/add/anonymous-99/K9/1

### Benchmark Applications  
1. [Robot Shop](https://github.com/instana/robot-shop)
2. [Custom Benchmark Application](https://github.com/resess/athena_kube_plugin/tree/master/kubernetes-client/test/endtoendtest/python)

### Load Generator
1.  [ApacheBench](https://httpd.apache.org/docs/2.4/programs/ab.html)
2.  [Apache JMeter](https://jmeter.apache.org/)


### Notes
File **kubernetes-client/scheduler.go** contains algorithm of our custom scheduler and below are description about the code: 
1. **curServices** and **curNodes** store the node and service information of the cluster.    
2. **func schedule()** is called periodically for re-scheduling. All scheduling code goes in this function.  
3. **CreateDeployment()**, **UpdateDeployment()**, **DeleteDeployment()** and **BindPodToNode()** make API calls to propagate scheduling decisions.    
4. **ListDeployment()** and **ListNode()** make API calls to get cluster status information.    
5. **generateMockData()** is used to generate mock data for testing.  
