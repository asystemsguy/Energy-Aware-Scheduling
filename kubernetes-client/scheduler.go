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
	"strings"
	"sort"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	clientmetricsV1 "github.com/prometheus/client_golang/api"

	log "github.com/sirupsen/logrus"

)

// define types
type ServiceIdentifier struct {
	Name string
	Namespace string
}

type NodeEntity struct {
	Id	int
	Name	string
	Cpu     float32
	Memory  float32
	Disk	float32
	Ports   []int
	Services []int
	Rack    string
	PM      string
}


type ServiceEntity struct {
	Id		int
	Name	ServiceIdentifier
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

type RackMapVal struct{
	pmName string
	ndList []NodeEntity
}

/* global variables */
// config items
const scheduleInterval = 60 // in seconds

const promeMonitorAdr = "http://10.42.0.65:9090"
const promeIstioAdr = "http://10.42.0.65:9090"

const NODECPU = "machine_cpu_cores"
const NODEMEM = "machine_memory_bytes"
const ISTIO_RCV = "istio_tcp_received_bytes_total"
const ISTIO_SENT = "istio_tcp_sent_bytes_total"

// API clients
var proAPI		metricsV1.API
var proAPI2		metricsV1.API
var clientSet		kubernetes.Interface

// Store cluster status
var curServices []ServiceEntity
var curNodes	[]NodeEntity
var masterNode = "rubin4"
var serviceMap  [][]int
var curClusters []Clusters
var rackMap map[string][]RackMapVal

var namespaceMap map[string][]ServiceEntity
var svcList	map[ServiceIdentifier]int

// Variables for getting next availble nodes in the rack hierarchy
// used in func getNextAvailableNode()
var curRack string
var curPM string
var curVM string

func startLogging() {

	//open a log file
	file, err := os.OpenFile("kube-scheduler.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }


    log.SetOutput(file)

    log.SetLevel(log.DebugLevel)
    log.SetReportCaller(true)

}


func main() {
	// Creating K8s interface

	startLogging()

	var kubeconfig *string

	// Find kubeconfig file in home directory to establish a connection interface
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
	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Initalize Monitoring Prometheus API client
	var prometheus_config clientmetricsV1.Config
	prometheus_config.Address = promeMonitorAdr
	prometheus_config.RoundTripper = clientmetricsV1.DefaultRoundTripper
	prometheus_client, err := clientmetricsV1.NewClient(prometheus_config)
	if err != nil {
		log.Fatal(err)
	}
	proAPI = metricsV1.NewAPI(prometheus_client)

	// Initalize Istio Prometheus API client
	var prometheus_config2 clientmetricsV1.Config
	prometheus_config2.Address = promeIstioAdr
	prometheus_config2.RoundTripper = clientmetricsV1.DefaultRoundTripper
	prometheus_client2, err := clientmetricsV1.NewClient(prometheus_config2)
	if err != nil {
	    log.Fatal(err)
			fmt.Println(err)
	}
	proAPI2 = metricsV1.NewAPI(prometheus_client2)


	// Stop the mother goroutine from ever diing
	stopCh := make(chan struct{})

	// generateMockData() // generate mock data
	curRack = "R1"
	curPM = "PM2"
	curVM = "rubin3"

	go Schedule()

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
func Update_clusters() {

	curServices, namespaceMap, svcList = ListPod()
	fmt.Println("-----------")
	fmt.Println("Current Services: ")
	for _, service := range curServices{
		log.Debug(service)
		fmt.Println(service)
	}
	fmt.Println("-----------")
	fmt.Println("Current Nodes: ")
	curNodes = ListNode()
	for _, node := range curNodes{
		log.Debug(node)
		fmt.Println(node)
	}
	fmt.Println("-----------")
	//generate Rack Map
	rackMap = generateRackMap()
	fmt.Println("Rack Map: ")
	for key,val := range rackMap{
		fmt.Println(key,val)
	}
	fmt.Println("-----------")
	//service mesh servicemap_generator
	serviceMap = servicemap_generator(namespaceMap)

	fmt.Println("Service Map:")
	fmt.Println(serviceMap)
	fmt.Println("-----------")
	//read the service map and create the services_bw.json
	mapServices := map[string]interface{}{"services":serviceMap}

	mapServicesB, _ := json.Marshal(mapServices)

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

    var service_clusters = result["service_clusters"].(map[string]interface{})

    var keys []int
    for k, _ := range service_clusters {
        var ki,_ = strconv.Atoi(k)
        keys = append(keys,ki )
    }
    sort.Ints(keys)
    curClusters = nil

	for _,key := range keys  {

        var tempCluster Clusters
        var temclusterProfile ServiceEntity
        var key_s = strconv.Itoa(key)
        var value = service_clusters[key_s]
        tempCluster.clusterid = key

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
        curClusters= append(curClusters, tempCluster)

	}
}

func Deploy_clusters() {

   	//Golang doesn't guareentee the interation is in order hence sort them
   	var keys []int
   	for k := range curClusters {
       	keys = append(keys, k)
   	}
   	sort.Ints(keys)

   	fmt.Println("Current clusters ",curClusters)
	// Find bins to place these clusters

   	for index_c,_ := range keys {
       		fmt.Println(index_c)
   	   	var cluster = curClusters[len(curClusters)-1-index_c]
       		if(!cluster.deployed) {
           		fmt.Println(cluster)
			availNode, pre := getNextAvailableNode(
				cluster.clusterProfile.Cpu, cluster.clusterProfile.Memory,
				cluster.clusterProfile.Disk, cluster.clusterProfile.Ports);
			if(pre){
					index_n := availNode.Id

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
					for _, ele := range cluster.services {
					   testBind(curServices[ele].Name.Name, curServices[ele].Name.Namespace, curNodes[index_n].Name);
					   fmt.Println("Bind ",curServices[ele].Name.Name, curServices[ele].Name.Namespace, curNodes[index_n].Name )
					}

			       	//Substract resources
					curNodes[index_n].Cpu -= cluster.clusterProfile.Cpu
					curNodes[index_n].Memory -= cluster.clusterProfile.Memory
					curNodes[index_n].Disk -= cluster.clusterProfile.Disk
					curNodes[index_n].Ports = append(curNodes[index_n].Ports,cluster.clusterProfile.Ports...)

					log.Debug("Cluster "+string(cluster.clusterid)+" deployed on node"+string(index_n))
					fmt.Println(curNodes)
					// update current VM, PM and rack
					curVM = availNode.Name
					curPM = availNode.PM
					curRack = availNode.Rack



	              //  goto Exit

	       	} else {
	       	    log.Debug("Cluster "+string(cluster.clusterid)+" cannot be deployed on any node")
	       	}
       		//Exit: //service is deployed move to next
		}
   }
}

func Schedule() {
	// Note: you can use curServices and curNodes which contains mock data
   for{
   	//Update the clusters by re-running clustering algo
	Update_clusters()

   	//check the resource availablity and deploy on kubernetes
  	Deploy_clusters()

   	log.Info("Done Scheduling ...\n")
   	fmt.Printf("Done Scheduling ...\n",)
   	time.Sleep(time.Second * scheduleInterval)
   }
}

// list pod function
func ListPod()(podList []ServiceEntity, nMap map[string][]ServiceEntity, svList map[ServiceIdentifier]int){

	pods, err := clientSet.CoreV1().Pods("").List(metav1.ListOptions{});
	if err != nil{
		panic(err)
	}

	size := len(pods.Items)
	nMap = make(map[string][]ServiceEntity, size)
	svList = make(map[ServiceIdentifier]int, size)

	for i, p := range pods.Items{
		cpuReq, memReq, diskReq, portReq, nameSpace := getPodCPUMemLim(&p)
		svcIdentifier := ServiceIdentifier{Name: p.Name, Namespace: nameSpace}
		serviceEnt := ServiceEntity{Id: i,
			Name: svcIdentifier,
			Cpu: cpuReq, Memory: memReq,
			Disk: diskReq, Ports: portReq}

		_, present := nMap[nameSpace]

		if present{
			nMap[nameSpace] = append(nMap[nameSpace], serviceEnt)
		} else{
			nMap[nameSpace] = []ServiceEntity{serviceEnt}
		}
		podList = append(podList,serviceEnt)

		svList[svcIdentifier] = i
	}
	return podList, nMap, svList
}

// list node function
func ListNode()(nodeList []NodeEntity){
	nodes, _ := clientSet.CoreV1().Nodes().List(metav1.ListOptions{})
	var nl []string

	for _, n := range nodes.Items{
		nl = append(nl, n.Name)
	}
	return getNodeCPUMem(nl)
}

// service map generator (Communication pattern Matrix)
func servicemap_generator(namespaceMap1 map[string][]ServiceEntity) ([][]int) {

    var namespace = "athena"
	ctx := context.Background()
    result_received, _ := proAPI2.Query(ctx,"istio_tcp_received_bytes_total", time.Now())

    s_received := strings.Split(result_received.String(), "]")
	result_sent, _ := proAPI2.Query(ctx,"istio_tcp_sent_bytes_total", time.Now())
    s_sent := strings.Split(result_sent.String(), "]")


	//get the list of pods in athena namespace and initilaize the length of the service map
	l := len(namespaceMap1[namespace])

    serviceMap_temp := make([][]int,l)
	for i:=0; i<l; i++ {
  	  serviceMap_temp[i] = make([]int, l)
	}

    serviceMap_temp2 := make([][]int,l)
	for i:=0; i<l; i++ {
  	  serviceMap_temp2[i] = make([]int, l)
	}
	podMap := make(map[string]int)
	fmt.Println(namespaceMap1[namespace])

	//create a map for pods and ids
	for j := 0; j < l; j++ {
		podMap[strings.Split(namespaceMap1[namespace][j].Name.Name, "-")[0]] = int(namespaceMap1[namespace][j].Id)
	}

	// retreive metrics for istio_tcp_received_bytes_total and store in the matrix
	for i := 0; i < (len(s_received)-1); i++ {

	    s1 := strings.Split( strings.Split( strings.Trim(s_received[i], "istio_tcp_received_bytes_total"), "@")[0], "=>")
		s2 := strings.Split(s1[0], ",")
		n,_ := strconv.Atoi(strings.Trim(s1[1], " "))
		index1 := strings.Trim(strings.Split(s2[7],"=")[1],"\"\"")
		index2 := strings.Trim(strings.Split(s2[15],"=")[1],"\"\"")
		if value1, exist1 := podMap[index1]; exist1{
			if value2, exist2:= podMap[index2]; exist2 {
				serviceMap_temp[value1][value2] += n
			}
		}
	}

	//retreive metrics from istio_tcp_sent_bytes_total and store in the matrix
 	for i := 0; i < (len(s_sent)-1); i++ {

	    s1 := strings.Split( strings.Split( strings.Trim(s_sent[i], "istio_tcp_sent_bytes_total"), "@")[0], "=>")
	 	s2 := strings.Split(s1[0], ",")
	 	n,_ := strconv.Atoi(strings.Trim(s1[1], " "))
		index1 := strings.Trim(strings.Split(s2[7],"=")[1],"\"\"")
		index2 := strings.Trim(strings.Split(s2[15],"=")[1],"\"\"")
		if value1, exist1 := podMap[index1]; exist1{
			if value2, exist2:= podMap[index2]; exist2 {
				serviceMap_temp[value1][value2] += n
			}
		}
	}

	//transform the pattern of the matrix for the clustering algorithm
	for i:= 0; i<l; i++ {
		for j:=0; j<l; j++{
			if ((serviceMap_temp[i][j] + serviceMap_temp[j][i]) == 0){
					serviceMap_temp2[i][j] = 1
				} else{
					serviceMap_temp2[i][j] = (serviceMap_temp[i][j] + serviceMap_temp[j][i])
			}

		}
	}

	return serviceMap_temp2
}
