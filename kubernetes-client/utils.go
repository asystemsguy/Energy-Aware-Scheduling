package main
import (
	"fmt"
	"bufio"
	"os"
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	k8stpv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/util/retry"
	apiresrc "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	log "github.com/sirupsen/logrus"

)

// get service by label function
func ListService(label string)([]int){
	port := []int{}
	//FIXME: Query only services with labelSelector or query this once
	services, _ := clientSet.CoreV1().Services("athena").List(metav1.ListOptions{})
	for _,n := range services.Items{
		if n.Spec.Selector["server"] == label{
			for _, port_s := range n.Spec.Ports{
				port = append(port, int(port_s.Port))		
			}
		}

	}
	return port
}

func getNextAvailableNode(cpu float32, memory float32, disk float32, ports []int)(NodeEntity, bool){
	pmList, pre1 := rackMap[curRack]
	pmNameVal := false
	if pre1 {
		
		for _, pm := range pmList{
			if pm.pmName == curPM{
				pmNameVal = true

				// check if current VM has enough capacities
				for _, vm := range pm.ndList{
					node := curNodes[vm.Id]
					if node.Name == curVM && node.Cpu >= cpu && 
					node.Memory >= memory && node.Disk >= disk &&
					!subset(ports, node.Ports) && node.Name != masterNode{
						return node,true; 
					}
				}

				// find VM in the same PM
				for _, vm := range pm.ndList{
					node := curNodes[vm.Id]
					if node.Name != curVM && node.Cpu >= cpu && 
					node.Memory >= memory && node.Disk >= disk &&
					!subset(ports, node.Ports) && node.Name != masterNode{
						return node,true; // find the next availble node in the same PM
					}
				}
			}
		}
		// find VM in the same rack
		if pmNameVal{
			for _,pm := range pmList{
				if pm.pmName != curPM{
					for _,vm := range pm.ndList{
						node := curNodes[vm.Id]
						if node.Name != curVM && node.Cpu >= cpu && 
						node.Memory >= memory && node.Disk >= disk &&
						!subset(ports, node.Ports) && node.Name != masterNode{
							return node,true;
						}
					}
				}
			}
		}else{
			fmt.Println("curPm name does not exist.")
		}
	}else{
		fmt.Println("curRack name does not exist.")
	}

	// find VM in other racks
	for rackNm,pmList := range rackMap{
		if rackNm != curRack{
			for _, pm := range pmList{
				for _, vm := range pm.ndList{
					node := curNodes[vm.Id]
					if node.Cpu >= cpu && node.Memory >= memory && 
						node.Disk >= disk && !subset(ports, node.Ports) &&
						node.Name != masterNode{
							return node,true; // find the next availble node in the same PM
					}
				}
			}
		}
	}

	// no node has enough capacities
	return NodeEntity{},false
}

// Rack map generator
func generateRackMap()(map[string][]RackMapVal){
	localRackMap := make(map[string][]RackMapVal)

	pmNdMap := make(map[string][]NodeEntity)
	rckPmMap := make(map[string][]string)
	recordPmSet := make(map[string]bool)

	for _, nodeEnt := range curNodes{
		// creating Rack and PM Map
		if _, value := recordPmSet[nodeEnt.PM]; !value {
			recordPmSet[nodeEnt.PM] = true
			_, p1 := rckPmMap[nodeEnt.Rack]
			if p1{
				rckPmMap[nodeEnt.Rack] = append(rckPmMap[nodeEnt.Rack], nodeEnt.PM)
			} else{
				rckPmMap[nodeEnt.Rack] = []string{nodeEnt.PM}
			}
		}

		//creat PM and Node Map
		_, p2 := pmNdMap[nodeEnt.PM]
		if p2{
			pmNdMap[nodeEnt.PM] = append(pmNdMap[nodeEnt.PM], nodeEnt)
		} else{
			pmNdMap[nodeEnt.PM] = []NodeEntity{nodeEnt}
		}

	}
	// fmt.Println("PM and Node map:", pmNdMap)
	// fmt.Println("Rack and PM map:", rckPmMap)

	for rackName, PmNames := range rckPmMap{
		localRackMap[rackName] = []RackMapVal{}
		for _,PmName := range PmNames{
			rackmapEntry := RackMapVal{pmName: PmName, ndList: pmNdMap[PmName]}
			localRackMap[rackName] = append(localRackMap[rackName], rackmapEntry)
		}
	}
	// fmt.Println("Rack Map:", localRackMap)
	return localRackMap
}

func getPodCPUMemReq(pod *apiv1.Pod) (float32, float32, float32, []int, string) {
	cpuReq := float32(0)
	memReq := float32(0)
	diskReq := float32(0)
	portReq := []int{}
	label := pod.Labels
	
	nameSpace := pod.Namespace
	for _, container := range pod.Spec.Containers {
		request := container.Resources.Requests
		cpuReqQuantity := request[apiv1.ResourceCPU]
		cpuReq += float32(cpuReqQuantity.MilliValue())/1000
		
		memReqQuantity := request[apiv1.ResourceMemory]
		memReqCont,_ := memReqQuantity.AsInt64()
		memReq += float32(memReqCont)
		
		diskReqQuantity := request[apiv1.ResourceStorage]
		disReqCont,_ := diskReqQuantity.AsInt64()
		diskReq += float32(disReqCont)
	}

	lb, p1 := label["server"]
	if p1{
		portReq = ListService(lb)
	}
	return cpuReq, memReq, diskReq, portReq, nameSpace
}

func getPodCPUMemLim(pod *apiv1.Pod) (float32, float32, float32, []int, string) {
	cpuLim := float32(0)
	memLim := float32(0)
	diskLim := float32(0)
	portLim := []int{}
	label := pod.Labels
	
	nameSpace := pod.Namespace
	for _, container := range pod.Spec.Containers {
		limit := container.Resources.Limits
		cpuLimQuantity := limit[apiv1.ResourceCPU]
		cpuLim += float32(cpuLimQuantity.MilliValue())/1000
		
		memLimQuantity := limit[apiv1.ResourceMemory]
		memLimCont,_ := memLimQuantity.AsInt64()
		memLim += float32(memLimCont)
		
		diskLimQuantity := limit[apiv1.ResourceStorage]
		disLimCont,_ := diskLimQuantity.AsInt64()
		diskLim += float32(disLimCont)
	}

	lb, p1 := label["server"]
	if p1{
		portLim = ListService(lb)
	}
	return cpuLim, memLim, diskLim, portLim, nameSpace
}


func getNodeCPUMem(nodeList []string)(nodeEntityList []NodeEntity){
	
	//nodePodMap := getPodNodeMap()
	// cpu query
	cpuMap := make(map[string]float64, len(nodeList))
	memMap := make(map[string]float64, len(nodeList))
	PmMap := make(map[string]string, len(nodeList))
	RackMap := make(map[string]string, len(nodeList))

	queryObj,_ := proAPI.Query(context.Background(), NODECPU, time.Now())
	samples := queryObj.ToSample()
	for _, s := range samples{
		//nm, present := s.Metric.GetValue("node")
		nm, present := s.Metric.GetValue("kubernetes_io_hostname")
		value := s.GetSampleValue()
		if !present{
			log.Info("cannot find query field")
			fmt.Println("cannot find query field")
		}else{
			cpuMap[nm] = value
		}
	}

	// memory query
	memRsp,_ := proAPI.Query(context.Background(), NODEMEM, time.Now())
	memSamples := memRsp.ToSample()
	for _, s := range memSamples{
		//nm, present := s.Metric.GetValue("node")
		nm, present := s.Metric.GetValue("kubernetes_io_hostname")
		pm, p1 := s.Metric.GetValue("PM")
		rack, p2 := s.Metric.GetValue("RACK")

		value := s.GetSampleValue()
		if !present{
			log.Info("cannot find query field")
			fmt.Println("cannot find query field")
		}else{
			memMap[nm] = value
			if !p1 || !p2 {
				fmt.Println("Cannt find pm and rack fields")
			}else{
				PmMap[nm] = pm
				RackMap[nm] = rack
			}
		}
	}
	
	for i, n := range nodeList{
		nodeEntityList = append(nodeEntityList, NodeEntity{
			Id: i, Name: n, 
			Cpu: float32(cpuMap[n]), 
			Memory: float32(memMap[n]),
			PM: PmMap[n],
			Rack: RackMap[n],})
			//Services: nodePodMap[n]})
	}
	return nodeEntityList
}

// Bind pod to node function
func BindPodToNode(clientSet kubernetes.Interface, bindInfo BindInfo) {
	err := clientSet.CoreV1().Pods(bindInfo.Namespace).Bind(&apiv1.Binding{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: bindInfo.Name,
			},
			Target: apiv1.ObjectReference{
				Namespace: bindInfo.Namespace,
				Name:      bindInfo.Nodename,
			}})
	if err != nil {
		fmt.Printf("Could not bind pod:%s to nodeName:%s, error: %v\n", bindInfo.Name, bindInfo.Nodename, err)
	}else{
		fmt.Printf("Successfully schedule pod:%s(%s), to nodeName:%s.\n",bindInfo.Name, bindInfo.Namespace, bindInfo.Nodename)
	}
}

//-------- Unused functions------------

// reference this when doing deployment related API calls
//deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
//ListDeployment(deploymentsClient)


const PODINFO = "kube_pod_info"

func getPodNodeMap()(pnMap map[string][]int){
	pnMap = make(map[string][]int)

	recordSet := make(map[string]bool)
	// pod info query
	queryObj,_ := proAPI.Query(context.Background(), PODINFO, time.Now())
	samples := queryObj.ToSample()
	for _,s := range samples{
		nodeNm, ps0 :=s.Metric.GetValue("node")
		podNm, ps1 :=s.Metric.GetValue("pod")
		podNs, ps2 :=s.Metric.GetValue("namespace")
		
		if _, value := recordSet[podNm]; !value {
			recordSet[podNm] = true
			if ps0 && ps1 && ps2 {
				// fmt.Printf("pod n: %s, ns: %s, node: %s", podNm, podNs, nodeNm)
				_, ps3 := pnMap[nodeNm]
				podId, ps4 := svcList[ServiceIdentifier{Name: podNm, Namespace: podNs}]
				if !ps4{
					fmt.Printf("Cannot find pod")
				}else if ps3{
					pnMap[nodeNm] = append(pnMap[nodeNm], podId)
				}else{
					pnMap[nodeNm] = []int{podId}
				}
			}
		}
	}
	return pnMap
}

func CreateDeployment(podName string, cpuLimit int64, deploymentsClient k8stpv1.DeploymentInterface){
	cpuQuantity := &apiresrc.Quantity{}
	cpuQuantity.Set(cpuLimit)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU: *cpuQuantity,
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
	return
}

func UpdateDeployment(podName string, deploymentsClient k8stpv1.DeploymentInterface){
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := deploymentsClient.Get(podName, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
		}

		result.Spec.Replicas = int32Ptr(1)                           // reduce replica count
		result.Spec.Template.Spec.Containers[0].Image = "nginx:1.13" // change nginx version
		_, updateErr := deploymentsClient.Update(result)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
	fmt.Println("Updated deployment...")
}

// functions to propagate scheduling decision to kubenertes cluster
func DeleteDeployment(podName string, deploymentsClient k8stpv1.DeploymentInterface){
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(podName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted deployment.")
}

func int32Ptr(i int32) *int32 { return &i }

// debug helper
func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

// update curServices with deployment info
// func ListDeployment(deploymentsClient k8stpv1.DeploymentInterface) (serviceList []ServiceEntity){
// 	fmt.Printf("Get deployment list \n");
// 	list, err := deploymentsClient.List(metav1.ListOptions{})
// 	if err != nil {
// 		panic(err)
// 	}
// 	for _, d := range list.Items {
// 		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas);
// 	}
// 	return
// } 
