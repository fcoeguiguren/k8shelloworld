package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	ctx := context.Background()
	figure.NewColorFigure("Hello World Demo", "", "green", true).Print()
	var kubeconfig *string
	var namespace string
	var podName string
	var filterLabel string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.StringVar(&namespace, "namespace", "samples", "namespace used to run this program")
	flag.StringVar(&podName, "podname", "helloworld", "name of the pod to be created")
	flag.StringVar(&filterLabel, "filterLabel", "k8s-app=kube-dns", "label to filter when listing pods")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	listNamespaces(ctx, clientset)
	createNamespace(ctx, clientset, namespace)
	time.Sleep(3 * time.Second)
	var labels = make(map[string]string)
	key := strings.Split(filterLabel, "=")
	labels[key[0]] = key[1]
	createPod(ctx, clientset, podName, namespace, labels)
	time.Sleep(3 * time.Second)
	listPodsWithLabels(ctx, clientset, filterLabel)
	deletePod(ctx, clientset, podName, namespace)
	time.Sleep(3 * time.Second)
	deleteNamespace(ctx, clientset, namespace)
	time.Sleep(3 * time.Second)
	listNamespaces(ctx, clientset)

}
func listPodsWithLabels(ctx context.Context, clientset *kubernetes.Clientset, filterLabel string) {
	fmt.Printf("=== Listing pods in all namspeaces with label %s  ===\n", filterLabel)
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		LabelSelector: filterLabel,
	})
	if err != nil {
		panic(err.Error())
	}
	for _, p := range pods.Items {
		fmt.Printf("Pod %s in namespace %s\n", p.Name, p.Namespace)
	}

}
func deletePod(ctx context.Context, clientset *kubernetes.Clientset, podName, namespace string) {
	fmt.Printf("=== Deleting %s pod in namespace %s ===\n", podName, namespace)
	err := clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("=== Pod deleted ===\n")
}
func createPod(ctx context.Context, clientset *kubernetes.Clientset, podName, namespace string, labels map[string]string) {
	fmt.Printf("=== Creating %s pod in namespace %s ===\n", podName, namespace)
	p := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "hellworld",
					Image: "k8s.gcr.io/echoserver:1.4",
				},
			},
		},
	}
	pod, err := clientset.CoreV1().Pods(namespace).Create(ctx, p, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("=== Pod %s created on %s  ===\n", pod.Name, &pod.ObjectMeta.CreationTimestamp)
}
func listNamespaces(ctx context.Context, clientset *kubernetes.Clientset) {
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("=== Found %d namespaces === \n", len(namespaces.Items))
	for _, s := range namespaces.Items {
		fmt.Println(s.Name)
	}
}
func createNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string) {
	fmt.Printf("=== Creating new namespace %s in the cluster ===\n", namespace)
	s := &core.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: namespace},
	}
	ns, err := clientset.CoreV1().Namespaces().Create(ctx, s, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("== Namespace %s created on %s ===\n", ns.Name, ns.ObjectMeta.CreationTimestamp)
}
func deleteNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string) {
	fmt.Printf("=== Deleting namespace %s ===\n", namespace)
	err := clientset.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("=== Namespace %s deleted ===\n", namespace)
}
