package k8s

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	twatch "k8s.io/client-go/tools/watch"
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

// tries to load the current workspace
func DeriveNamespace(namespace string) string {
	if namespace != "" {
		return namespace
	}
	clientCfg, _ := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	namespace = clientCfg.Contexts[clientCfg.CurrentContext].Namespace
	if namespace == "" {
		return "default"
	}
	return namespace
}

func GuessPodNameFromHost(hostname string) (string, error) {
	return strings.Split(hostname, ".")[0], nil
}

func KeepPodDead(ctx context.Context, clientset *kubernetes.Clientset, name, namespace string, grace int64, done chan error, pq chan map[string]string) {
	cl := clientset.CoreV1().Pods(namespace)
	// check the pod exists
	pod, err := cl.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		done <- fmt.Errorf("can't get the pod %s in %s; got %s", name, namespace, err)
	}
	// setup watch & deletion
	wf := func(ctx context.Context, options metav1.ListOptions) (watch.Interface, error) {
		t := int64(60)
		return cl.Watch(ctx, metav1.ListOptions{TimeoutSeconds: &t})
	}
	deletePod := func(failOnNotFound bool) {
		pq <- map[string]string{
			"event":     "deleting pod",
			"name":      name,
			"namespace": namespace,
		}
		err := cl.Delete(ctx, name, *metav1.NewDeleteOptions(grace))
		if errors.IsNotFound(err) {
			pq <- map[string]string{
				"event":     "pod not found",
				"name":      name,
				"namespace": namespace,
			}
			if failOnNotFound {
				done <- fmt.Errorf("pod %s in %s not found; got %s", name, namespace, err)
			}
		} else if err != nil {
			done <- fmt.Errorf("error deleting pod; got %s", err)
		}
	}
	wr, err := twatch.NewRetryWatcherWithContext(ctx, pod.ResourceVersion, &cache.ListWatch{WatchFuncWithContext: wf})
	if err != nil {
		done <- fmt.Errorf("can't create retry watcher; got %s", err)
	}
	// do the initial delete
	deletePod(true)
	// watch for the pod getting re-added & delete again
	for event := range wr.ResultChan() {
		item := event.Object.(*corev1.Pod)
		if item.Name != name {
			continue
		}
		switch event.Type {
		case watch.Added:
			pq <- map[string]string{
				"debug":     "true",
				"event":     "pod event",
				"type":      string(event.Type),
				"name":      item.Name,
				"namespace": item.Namespace,
			}
			deletePod(false)
		}
	}
}
