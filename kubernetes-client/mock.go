package main
import (
	"fmt"
	log "github.com/sirupsen/logrus"
)
//const promeMonitorAdr = "http://prometheus.istio-system.svc.cluster.local:9090"
//const promeIstioAdr = "http://prometheus.istio-system.svc.cluster.local:9090"

//const NODECPU = "kube_node_status_capacity_cpu_cores"
//const NODEMEM = "kube_node_status_capacity_memory_bytes"


var promeQueryPara = [...]string{
	"kube_node_labels",
	"kube_node_status_capacity_cpu_cores",
	"kube_node_status_capacity_memory_bytes",
	"kube_pod_info",
	"kube_pod_container_resource_limits_cpu_cores",
	"kube_pod_container_resource_limits_memory_bytes",
	"kube_pod_container_resource_requests_cpu_cores",
	"kube_pod_container_resource_requests_memory_bytes",
	// "istio_tcp_received_bytes_total",
	// "istio_tcp_sent_bytes_total",
	}

var nodeNN="rubin4"
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

func testBind(podName, nameSpace, nodeName string){
	bindEntity := BindInfo{Name: podName, Namespace: nameSpace, Nodename: nodeName}
	BindPodToNode(clientSet, bindEntity)
}

func testing_clusters(){
	curServices, namespaceMap, svcList = ListPod()
        fmt.Println("Current Services: ")
        for _, service := range curServices{
                log.Debug(service)
                fmt.Println(service)
		if service.Name.Namespace == "athena" {
			fmt.Println(service)
			testBind(service.Name.Name, service.Name.Namespace, nodeNN)
		}
        }
}

func test_getNextAvailableNode(){
	curRack = "rack1"
	curPM = "PM1"
	curVM = "rubin2"
	rackMap = make(map[string][]RackMapVal)
	rackMap["rack1"] = []RackMapVal{
		RackMapVal{pmName:"PM1", ndList:[]NodeEntity{
			NodeEntity{Name: "rubin1", Cpu:1.0, Memory: 1024, Disk: 0.0 },
			NodeEntity{Name: "rubin2", Cpu:2.5, Memory: 512, Disk: 0.0 },
			NodeEntity{Name: "rubin3", Cpu:2.5, Memory: 125, Disk: 0.0 },},
		},
		RackMapVal{pmName:"PM2", ndList:[]NodeEntity{
			NodeEntity{Name: "rubin4", Cpu:1.0, Memory: 1024, Disk: 0.0 },
			NodeEntity{Name: "rubin5", Cpu:2.5, Memory: 512, Disk: 0.0 },},
		},
	}
	rackMap["rack2"] = []RackMapVal{
		RackMapVal{pmName:"PM3", ndList:[]NodeEntity{
			NodeEntity{Name: "rubin6", Cpu:9.0, Memory: 1024, Disk: 0.0 },
			NodeEntity{Name: "rubin7", Cpu:2.5, Memory: 512, Disk: 0.0 },
			NodeEntity{Name: "rubin8", Cpu:2.5, Memory: 125, Disk: 0.0 },},
		},
		RackMapVal{pmName:"PM4", ndList:[]NodeEntity{
			NodeEntity{Name: "rubin9", Cpu:1.0, Memory: 1024, Disk: 0.0 },
			NodeEntity{Name: "rubin10", Cpu:2.5, Memory: 512, Disk: 0.0 },},
		},
	}
	fmt.Println(getNextAvailableNode(1.25, 215, 0, []int{}))
	fmt.Println(getNextAvailableNode(2.5, 0, 0, []int{}))
	fmt.Println(getNextAvailableNode(5, 0, 0, []int{}))
	fmt.Println(getNextAvailableNode(0, 5, 0, []int{}))
	fmt.Println(getNextAvailableNode(1000, 5, 0, []int{}))
	
	curVM = "ruby"
	curPM = "PM"
	fmt.Println(getNextAvailableNode(1, 0, 0, []int{}))
	
}