package k8s

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	toolsWatch "k8s.io/client-go/tools/watch"
)

func GetClient(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func GuessPodNameFromHost(hostname string) (string, error) {
	return strings.Split(hostname, ".")[0], nil
}

func KeepPodDead(clientset *kubernetes.Clientset, ctx context.Context, name, namespace string, done chan error, pq chan map[string]string) {
	cl := clientset.CoreV1().Pods(namespace)
	wf := func(ctx context.Context, options metav1.ListOptions) (watch.Interface, error) {
		t := int64(60)
		return cl.Watch(ctx, metav1.ListOptions{TimeoutSeconds: &t})
	}
	wr, err := toolsWatch.NewRetryWatcherWithContext(ctx, "1", &cache.ListWatch{WatchFuncWithContext: wf})
	if err != nil {
		done <- fmt.Errorf("can't create retry watcher; got %s", err)
	}
	for event := range wr.ResultChan() {
		item := event.Object.(*corev1.Pod)
		if item.Name != name {
			continue
		}
		pq <- map[string]string{
			"event":     "pod event",
			"type":      string(event.Type),
			"name":      item.Name,
			"namespace": item.Namespace,
		}
		switch event.Type {
		case watch.Modified:
		case watch.Added:
			pq <- map[string]string{
				"event":     "deleting pod",
				"name":      item.Name,
				"namespace": item.Namespace,
			}
			err = cl.Delete(ctx, name, *metav1.NewDeleteOptions(0))
			if err != nil {
				done <- fmt.Errorf("error deleting pod; got %s", err)
			}
		}
	}
}
