package main
import (
	"fmt"
	"bufio"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	k8stpv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/util/retry"
	apiresrc "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)
// reference this when doing deployment related API calls
//deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
//ListDeployment(deploymentsClient)
// extract CPU and memory from Pod specification

func getCPUMemRequest(pod *apiv1.Pod) (float32, float32) {
	cpuReq := float32(0)
	memReq := float32(0)
	for _, container := range pod.Spec.Containers {
		request := container.Resources.Requests
		cpuReqQuantity := request[apiv1.ResourceCPU]
		cpuReq += float32(cpuReqQuantity.MilliValue())/1000
		memReqQuantity := request[apiv1.ResourceMemory]
		memReqCont,_ := memReqQuantity.AsInt64()
		memReq += float32(memReqCont)
	}
	//fmt.Printf("cpuReq %f, memReq %f\n", cpuReq, memReq);
	return cpuReq, memReq
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

func BindPodToNode(clientSet kubernetes.Interface, bindInfo BindInfo) {
	for {
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
			fmt.Printf("Could not bind pod:%s to nodeName:%s, error: %v", bindInfo.Name, bindInfo.Nodename, err)
		}
	}
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