package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.io/lmxia/hybrid-gaia/pkg"
)

var (
	kubeconfigGlobal string
	kubeconfigLocal  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "hybrid-gaia",
		Short: "A tool to manage gaia desc and deploy across Kubernetes clusters",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create CRDs and Deployments",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pkg.CreateResource(kubeconfigGlobal, kubeconfigLocal); err != nil {
				log.Fatalf("Error executing create command: %v", err)
			}
		},
	}

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear Description and Deployments",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pkg.ClearResource(kubeconfigGlobal, kubeconfigLocal); err != nil {
				log.Fatalf("Error executing clear command: %v", err)
			}
		},
	}
	rootCmd.AddCommand(createCmd, clearCmd)

	createCmd.Flags().StringVar(&kubeconfigGlobal, "kubeconfig-global", "/Users/xialingming/.kube/scn_global.yaml", "Path to the kubeconfig file for cluster global")
	createCmd.Flags().StringVar(&kubeconfigLocal, "kubeconfig-local", "/Users/xialingming/.kube/scn_tencent_cluster.conf", "Path to the kubeconfig file for cluster local")

	clearCmd.Flags().StringVar(&kubeconfigGlobal, "kubeconfig-global", "/Users/xialingming/.kube/scn_global.yaml", "Path to the kubeconfig file for cluster global")
	clearCmd.Flags().StringVar(&kubeconfigLocal, "kubeconfig-local", "/Users/xialingming/.kube/scn_tencent_cluster.conf", "Path to the kubeconfig file for cluster local")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
