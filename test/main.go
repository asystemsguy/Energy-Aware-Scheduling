package main

import (
	"fmt"
	"strconv"
	"flag"
	"os"
	"os/exec"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"context"
	"time"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	clientmetricsV1 "github.com/prometheus/client_golang/api"

	log "github.com/sirupsen/logrus"

	// "k8s.io/client-go/rest" 
)

// define types
type NodeEntity struct {
	Id 		int
	Name 	string
	Cpu     float32
	Memory  float32
	Disk	float32
	Ports   []int
	Services []int
}

type ServiceEntity struct {
	Id		int
	Name	string
	Cpu     float32
	Memory  float32
	Disk	float32
	Ports   []int
	deployed bool
}

type BindInfo struct {
	Name      string
	Namespace string
	Nodename  string
}

type Clusters struct {
    clusterid int
    services []int
    clusterProfile ServiceEntity
    deployed bool
}

/* global variables, bad practice but use it for now */
// store cluster status

var curServices []ServiceEntity
var curNodes	[]NodeEntity
var serviceMap  [][]int
var curClusters []Clusters


func startLogging() {

	//open a log file
	file, err := os.OpenFile("kube-scheduler.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }

    //log.SetOutput(os.Stdout)
    log.SetOutput(file)
    //log.SetFormatter(&log.JSONFormatter{})
    log.SetLevel(log.DebugLevel)
    log.SetReportCaller(true)
	
}


func main() {
	// Creating K8s interface
	// CONTAINER SETTING: In-cluster
	// config, err := rest.InClusterConfig()

	startLogging()

	var kubeconfig *string
	
	//find kubeconfig file in home directory to establish a connection interface
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// Get current services and nodes, update service and node lists
	curServices = ListPod(clientset)
	for _, service := range curServices{
		log.Debug("Service id: " + string(service.Id) + "name: " + service.Name, service.Cpu, service.Memory)
		//fmt.Println("Service id: " + string(service.Id) + "name: " + service.Name, service.Cpu, service.Memory)
	}
	
	curNodes = ListNode(clientset)
	for _, node := range curNodes{
		log.Debug("Node id: " + string(node.Id) + "name: " + node.Name)
		//fmt.Println("Node id: " + string(node.Id) + "name: " + node.Name)
	}

	var prometheus_config clientmetricsV1.Config

	prometheus_config.Address = "http://10.44.0.2:9090"
	prometheus_config.RoundTripper = clientmetricsV1.DefaultRoundTripper

	prometheus_client, err := clientmetricsV1.NewClient(prometheus_config)
	if err != nil {
	    log.Fatal(err)
	}
	
	proAPI := metricsV1.NewAPI(prometheus_client)

  	log.Debug(proAPI.Query(ctx,"kube_node_labels",time.Now()))

	// Stop the mother goroutine from ever diing
	stopCh := make(chan struct{}) 
	
	generateMockData() // generate mock data

	go schedule() // TODO(Wendy): call this periodically 
	
	//defer file.Close()

	<-stopCh
}

func subset(subsetArr, setArr []int) bool {

    set := make(map[int]int)
    for _, value := range setArr {
        set[value] += 1
    }

    for _, value := range subsetArr {
        if count, found := set[value]; !found {
            return false
        } else if count < 1 {
            return false
        } else {
            set[value] = count - 1
        }
    }

    return true
}



// Executes the python clustering script and loads the new clusters
func update_clusters() {

   //read the service map and create the services_bw.json 

   mapServices := map[string]interface{}{"services":serviceMap}
	
   mapServicesB, _ := json.Marshal(mapServices)

  // log.Debug(string(mapServicesB))
   
   err := ioutil.WriteFile("services_bw.json", mapServicesB, 0644)
  
	if err != nil {
	    log.Fatal(err)
	}


   //call high.py to start clustering and return service_clusters.json file
    
    _ , err = exec.Command("python3", "high.py").Output()

    if err != nil {
    	log.Fatal(err)
	}

   //get all the clusters from service_clusters.json file

    ServiceClustersFile, err := os.Open("service_clusters.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Successfully Opened service_clusters.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer ServiceClustersFile.Close()

    byteValue, _ := ioutil.ReadAll(ServiceClustersFile)

	var result map[string]interface{}

	json.Unmarshal([]byte(byteValue), &result)

	log.Debug(result["service_clusters"])

	for key, value := range result["service_clusters"].(map[string]interface{}) {


         var tempCluster Clusters
         var temclusterProfile ServiceEntity
         tempCluster.clusterid, err = strconv.Atoi(key)

         temclusterProfile.Cpu =  0
         temclusterProfile.Memory =  0
         temclusterProfile.Disk =  0
         temclusterProfile.Ports = append(temclusterProfile.Ports,-1)
         
         if err != nil {
			log.Fatal(err)
		 }

         for _, data := range value.([]interface{}) {
         	 
         	 data_i, _ := data.(float64)
         	 tempCluster.services = append(tempCluster.services, int(data_i))
         	 
         	 temclusterProfile.Cpu += curServices[int(data_i)].Cpu
         	 temclusterProfile.Memory += curServices[int(data_i)].Memory
         	 temclusterProfile.Disk += curServices[int(data_i)].Disk
         	 temclusterProfile.Ports = append(temclusterProfile.Ports,curServices[int(data_i)].Ports...)
         	 tempCluster.deployed = false
         }
  		
  		 tempCluster.clusterProfile = temclusterProfile
         curClusters = append(curClusters, tempCluster)
	}
	log.Debug(curClusters)	

}

func deploy_clusters() {

   // log.Error(curClusters)
	// Find bins to place these clusters
   for index_c, _ := range curClusters {

   	   var cluster = curClusters[len(curClusters)-1-index_c]

       if(!cluster.deployed) {
	       for index_n, node := range curNodes {

	       	    if(cluster.clusterProfile.Cpu <= node.Cpu && cluster.clusterProfile.Memory <= node.Memory && cluster.clusterProfile.Disk <= node.Disk && !subset(cluster.clusterProfile.Ports,node.Ports)) {

	       	    	//Deploy the cluster on the node
	       	    	// Remove all children 
	       	    	for index_o, cluster_o := range curClusters {
	       	    		if(!cluster_o.deployed) {
		       	    		if(subset(cluster_o.services,cluster.services)) {

		       	    			curClusters[index_o].deployed = true

		       	    		}
		       	    	}
	       	    	}

	       	    	//Tell to kubernetes to deploy 
	       	    	
	       	    	


	       	    	//Substract resources
	                curNodes[index_n].Cpu -= cluster.clusterProfile.Cpu
	                curNodes[index_n].Memory -= cluster.clusterProfile.Memory  
	                curNodes[index_n].Disk -= cluster.clusterProfile.Disk  
	                curNodes[index_n].Ports = append(curNodes[index_n].Ports,cluster.clusterProfile.Ports...)

	                log.Debug("Cluster "+string(cluster.clusterid)+" deployed on node"+string(index_n))

	                goto Exit

	       	    }  else {

	       	    	log.Debug("Cluster "+string(cluster.clusterid)+" not deployed on node"+string(index_n))

	       	    }
	       }

	       Exit: //service is deployed move to next
	    }
   }
}

// this function is called every hour
func schedule() {
	// TODO(Harsha): add scheduing code here 
	// Note: you can use curServices and curNodes which contains mock data
   
   //Update the clusters by re-running clustering algo
   update_clusters()

   // TODO(Harsha): deploy only when clusters have changed
   //check the resource availablity and deploy on kubernetes
   deploy_clusters()

   log.Info("Done Scheduling ...\n")
}

//list pod function 
func ListPod(clientSet kubernetes.Interface)(podList []ServiceEntity){
	pods, err := clientSet.CoreV1().Pods("").List(metav1.ListOptions{});
	if err != nil{
		panic(err)
	}

	for i, p := range pods.Items{
		cpuReq, memReq := getCPUMemRequest(&p)
		podList = append(podList,ServiceEntity{Id: i, Name: p.Name, Cpu: cpuReq, Memory: memReq})
	}

	return podList
}

//list node function 
func ListNode(clientSet kubernetes.Interface)(nodeList []NodeEntity){
	nodes, _ := clientSet.CoreV1().Nodes().List(metav1.ListOptions{})

	for i, n := range nodes.Items{
		nodeList = append(nodeList, NodeEntity{Id: i, Name:n.Name})
	}

	return nodeList
}

func testBind(clientSet kubernetes.Interface, podName, nameSpace, nodeName string){
	bindEntity := BindInfo{Name: podName, Namespace: nameSpace, Nodename: nodeName}
	BindPodToNode(clientSet, bindEntity)

}

// test helper
func generateMockData(){
	
	log.Error("Data is mocked\n")

	curServices = [] ServiceEntity{ 
		ServiceEntity{Id: 1, Cpu: 1.0, Memory: 256, Disk: 4096, Ports: []int{3030, 8080},deployed: false},
		ServiceEntity{Id: 2, Cpu: 1.0, Memory: 512, Disk: 4096, Ports: []int{4030, 22}, deployed: false},
		ServiceEntity{Id: 3, Cpu: 1.5, Memory: 125, Disk: 4096, Ports: []int{2022, 3000}, deployed: false},
		ServiceEntity{Id: 4, Cpu: 2.0, Memory: 256, Disk: 4096, Ports: []int{30, 3040}, deployed: false},
		ServiceEntity{Id: 5, Cpu: 2.0, Memory: 256, Disk: 4096, Ports: []int{30, 3040}, deployed: false},
	}
	
	curNodes = [] NodeEntity{
		NodeEntity{Id: 1, Cpu: 14.0, Memory: 8096, Disk: 10096, Ports: []int{}, Services: []int{}},
		NodeEntity{Id: 2, Cpu: 15.0, Memory: 13096, Disk: 13096, Ports: []int{}, Services: []int{}},
	}
    
	serviceMap = [][]int{
	{1,9,6,1,1},
	{9,1,1,1,5},
	{6,1,1,1,1},
	{1,1,1,1,8},
	{1,5,1,8,1},
	}	

	fmt.Println("service map:", serviceMap)

}
