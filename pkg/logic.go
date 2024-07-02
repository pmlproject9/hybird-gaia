package pkg

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	crdv1 "github.com/lmxia/gaia/pkg/apis/apps/v1alpha1"
	clientset "github.com/lmxia/gaia/pkg/generated/clientset/versioned"
)

func CreateResource(kubeconfigGlobal, kubeconfigLocal string) error {
	// 读取 Description CRD 文件
	descriptionData, err := ioutil.ReadFile("/Users/xialingming/lmxia/ar-demo-desc.yaml")
	if err != nil {
		return fmt.Errorf("failed to read Description file: %w", err)
	}

	// 创建 A 集群客户端
	configGlobal, err := clientcmd.BuildConfigFromFlags("", kubeconfigGlobal)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig for cluster A: %w", err)
	}
	clientsetGlobal, err := clientset.NewForConfig(configGlobal)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client for cluster A: %w", err)
	}

	// 在 A 集群中创建 Description CRD
	desc, errCreate := createDescription(clientsetGlobal, descriptionData)
	if errCreate != nil {
		return fmt.Errorf("failed to create Description CRD in cluster A: %w", errCreate)
	}

	// 创建 B 集群客户端
	configLocal, err := clientcmd.BuildConfigFromFlags("", kubeconfigLocal)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig for cluster B: %w", err)
	}
	clientsetB, err := kubernetes.NewForConfig(configLocal)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client for cluster B: %w", err)
	}

	// 在 B 集群中创建两个 Deployment
	err = createDeployments(clientsetB, "/Users/xialingming/lmxia/demo-deploys")
	if err != nil {
		return fmt.Errorf("failed to create Deployments in cluster B: %w", err)
	}

	// 定时查询 A 集群中的 resourcebinding CRD 状态
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// 创建通道用于用户输入
	inputChan := make(chan int)

	// 启动协程读取用户输入
	go func() {
		for {
			var choice int
			fmt.Println("Enter your choice:")
			_, err := fmt.Scan(&choice)
			if err == nil {
				inputChan <- choice
			} else {
				fmt.Println("Invalid input. Please enter a number.")
			}
			return
		}
	}()

	var resourceBindings []crdv1.ResourceBinding

	for {
		select {
		case <-ticker.C:
			resourceBindings, err = queryResourceBindings(clientsetGlobal, desc.Name)
			if err != nil {
				log.Printf("failed to query resourcebindings: %v", err)
				continue
			}

			clearScreen()
			for i, rb := range resourceBindings {
				fmt.Printf("[%d] %s\n", i, rb.Name)
			}

		case choice := <-inputChan:
			if choice < 0 || choice >= len(resourceBindings) {
				fmt.Println("Invalid choice")
				continue
			}

			// 更新选中的 resourcebinding CRD 的状态
			err = updateResourceBindingStatus(clientsetGlobal, resourceBindings[choice])
			if err != nil {
				log.Printf("failed to update resourcebinding: %v", err)
			}
			fmt.Printf("You Have Selected Resourcebinding: %s\n", resourceBindings[choice].Name)
			return nil
		}
	}
}

func ClearResource(kubeconfigGlobal, kubeconfigLocal string) error {
	// 清理 A 集群中的 Description CRD
	configGlobal, err := clientcmd.BuildConfigFromFlags("", kubeconfigGlobal)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig for cluster A: %w", err)
	}
	clientsetGlobal, err := clientset.NewForConfig(configGlobal)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client for cluster A: %w", err)
	}

	// 读取 Description CRD 文件
	descriptionData, err := ioutil.ReadFile("/Users/xialingming/lmxia/ar-demo-desc.yaml")
	if err != nil {
		return fmt.Errorf("failed to read Description file: %w", err)
	}

	err = deleteDescriptionCRD(clientsetGlobal, descriptionData)
	if err != nil {
		return fmt.Errorf("failed to delete Description CRD in cluster A: %w", err)
	}

	// 清理 B 集群中的 Deployments
	configLocal, err := clientcmd.BuildConfigFromFlags("", kubeconfigLocal)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig for cluster B: %w", err)
	}
	clientsetLocal, err := kubernetes.NewForConfig(configLocal)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client for cluster B: %w", err)
	}

	err = deleteDeployments(clientsetLocal, "/Users/xialingming/lmxia/demo-deploys")
	if err != nil {
		return fmt.Errorf("failed to delete Deployments in cluster B: %w", err)
	}

	return nil
}
