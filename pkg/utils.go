package pkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/lmxia/gaia/pkg/common"
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	crdv1 "github.com/lmxia/gaia/pkg/apis/apps/v1alpha1"
	clientset "github.com/lmxia/gaia/pkg/generated/clientset/versioned"
)

func createDescription(clientset *clientset.Clientset, crdData []byte) (*crdv1.Description, error) {
	var description crdv1.Description
	err := yaml.Unmarshal(crdData, &description)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Description CRD: %w", err)
	}

	desc, errCreate := clientset.AppsV1alpha1().Descriptions(description.Namespace).Create(context.TODO(), &description, metav1.CreateOptions{})
	if errCreate != nil {
		return nil, fmt.Errorf("failed to create Description CRD: %w", err)
	}

	return desc, nil
}

func createDeployments(clientset *kubernetes.Clientset, path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		deployData, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", path, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read deployment file %s: %w", file.Name(), err)
		}

		var deployment v1.Deployment
		err = yaml.Unmarshal(deployData, &deployment)
		if err != nil {
			return fmt.Errorf("failed to unmarshal deployment file %s: %w", file.Name(), err)
		}

		_, err = clientset.AppsV1().Deployments(deployment.Namespace).Create(context.TODO(), &deployment, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create deployment %s: %w", file.Name(), err)
		}
	}

	return nil
}

func queryResourceBindings(clientset *clientset.Clientset, descName string) ([]crdv1.ResourceBinding, error) {
	rbs, err := clientset.AppsV1alpha1().ResourceBindings(common.GaiaRBMergedReservedNamespace).List(context.TODO(),
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("apps.gaia.io/ori.desc.name=%s", descName),
		})
	if err != nil {
		return nil, fmt.Errorf("failed to list resourcebindings: %w", err)
	}

	return rbs.Items, nil
}

func updateResourceBindingStatus(clientset *clientset.Clientset, rb crdv1.ResourceBinding) error {
	// 在这里实现更新 resourcebinding CRD 状态的逻辑
	rb.Spec.StatusScheduler = crdv1.ResourceBindingSelected
	_, err := clientset.AppsV1alpha1().ResourceBindings(common.GaiaRBMergedReservedNamespace).Update(context.TODO(),
		&rb, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update resourcebindings status: %w", err)
	}
	return nil
}

func deleteDescriptionCRD(clientset *clientset.Clientset, crdData []byte) error {
	var description crdv1.Description
	err := yaml.Unmarshal(crdData, &description)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Description CRD: %w", err)
	}

	err = clientset.AppsV1alpha1().Descriptions(description.Namespace).Delete(context.TODO(), description.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete Description CRD: %w", err)
	}

	return nil
}

func deleteDeployments(clientset *kubernetes.Clientset, path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		deployData, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", path, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read deployment file %s: %w", file.Name(), err)
		}

		var deployment v1.Deployment
		err = yaml.Unmarshal(deployData, &deployment)
		if err != nil {
			return fmt.Errorf("failed to unmarshal deployment file %s: %w", file.Name(), err)
		}

		err = clientset.AppsV1().Deployments(deployment.Namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to create deployment %s: %w", file.Name(), err)
		}
	}

	return nil
}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
