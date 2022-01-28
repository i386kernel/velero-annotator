package main

import (
	"context"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	var kubeconfig *string
	var namespace = "akme"

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional "+
			"absolute path to kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {
		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		for _, item := range pods.Items {
			for _, vols := range item.Spec.Volumes {
				if strings.Contains(vols.Name, "-api-") || (vols.VolumeSource.EmptyDir.Size() > 1) {
					continue
				}
				fmt.Printf("Pod: %s --- > VolumeName: %s", item.Name, vols.Name)
				annotations := map[string]string{"backup.velero.io/backup-volumes":vols.Name}
				item.SetAnnotations (annotations)
				_, err := clientset.CoreV1().Pods(namespace).Update(context.TODO(), &item, metav1.UpdateOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}
