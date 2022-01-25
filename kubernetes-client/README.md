
### Documentation  
* This plugin is running outside the cluster. It schedules pods that are pending periodically according to historical network statistics.
* cluster scheduling algorithm code is implemented in high.py and API interface is implemented in scheduler.go and utils.go.
* The plugin uses IP address to establish connection with Prometheus server. The IP address is specified in **const promeMonitorAdr**  
* The plugin connects to Kubernetes API server by reading Kubernetes config file and uses go-client api.  
* **curServices** and **curNodes** store the node and service information of the cluster.    
* **schedule()** is called periodically for re-scheduling. All scheduling code goes in this function.  
* **getPodCPUMemLim()** is used to get user-defined memory and cpu limits for pod 
* **getNodeCPUMem()** is used to get cpu and memory capacity of nodes through Prometheus
* **generateMockData()** is used to generate mock data for testing.  
