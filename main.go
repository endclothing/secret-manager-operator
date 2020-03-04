package main

import (
	"context"
	"fmt"
	"log"
	"time"

	types "github.com/endclothing/secret-manager-operator/pkg/apis/endclothing.com/v1"
	v1 "github.com/endclothing/secret-manager-operator/pkg/client/clientset/versioned/typed/endclothing.com/v1"

	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"

	corev1 "k8s.io/api/core/v1"
	errorsv1 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"net/http"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Couldn't retrieve in-cluster Kubernetes client config: %+v", err)
	}

	msClient, err := v1.NewForConfig(config)
	if err != nil {
		log.Fatalf("Couldn't construct CRD client: %+v", err)
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Couldn't construct Kubernetes API client: %+v", err)
	}
	ctx := context.Background()
	secretManagerClient, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("Couldn't construct SecretManager client: %+v", err)
	}

	_, msController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return msClient.ManagedSecrets("").List(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return msClient.ManagedSecrets("").Watch(lo)
			},
		},
		&types.ManagedSecret{},
		15*time.Second,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				asms, ok := obj.(*types.ManagedSecret)
				if !ok {
					log.Printf("This event isn't a ManagedSecret, passing on it")
					return
				}
				go reconcileReference(msClient, k8sClient, secretManagerClient, asms)
			},
			UpdateFunc: func(_, obj interface{}) {
				asms, ok := obj.(*types.ManagedSecret)
				if !ok {
					log.Printf("This event isn't a ManagedSecret, passing on it")
					return
				}
				go reconcileReference(msClient, k8sClient, secretManagerClient, asms)
			},
		},
	)

	go msController.Run(wait.NeverStop)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Help, I'm alive!"))
	})
	log.Printf("ManagedSecret operator starting")
	http.ListenAndServe(":8080", nil)
}

func reconcileReference(msClient *v1.ComV1Client, k8sClient *kubernetes.Clientset, secretManagerClient *secretmanager.Client, ms *types.ManagedSecret) {
	// Handle deletions
	if !ms.ObjectMeta.DeletionTimestamp.IsZero() {
		finalize(msClient, k8sClient, ms)
		return
	}

	changed, err := reconcileReferenceHelper(msClient, k8sClient, secretManagerClient, ms)
	// UpdateFunc is called even if the object hasn't changed, so bail early if the managed secret has been reconciled already
	// to avoid spamming
	if !changed {
		return
	}
	if err != nil {
		ms.Status.Error = err.Error()
		log.Printf("Error reconciling ManagedSecret %s: %+v", ms.String(), err)
	} else {
		ms.Status.Project = ms.Spec.Project
		ms.Status.Secret = ms.Spec.Secret
		ms.Status.Generation = ms.Spec.Generation
		ms.Status.Error = ""
		log.Printf("Successfully reconciled ManagedSecret %s", ms.String())
	}
	_, err = msClient.ManagedSecrets(ms.Namespace).Update(ms)
	if err != nil {
		log.Printf("Error updating ManagedSecret %s status: %+v", ms.String(), err)
	}
}

func reconcileReferenceHelper(msClient *v1.ComV1Client, k8sClient *kubernetes.Clientset, secretManagerClient *secretmanager.Client, ms *types.ManagedSecret) (bool, error) {
	if !isStale(ms) {
		return false, nil
	}

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%d", ms.Spec.Project, ms.Spec.Secret, ms.Spec.Generation),
	}
	ctx := context.Background()
	secret, err := secretManagerClient.AccessSecretVersion(ctx, req)
	if err != nil {
		return true, err
	}

	k8sSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s", ms.Spec.Secret),
			Namespace: ms.Namespace,
		},
		Data: map[string][]byte{
			"contents": []byte(secret.Payload.Data),
		},
		Type: "Opaque",
	}
	err = createOrUpdateSecret(k8sClient, k8sSecret)
	if err != nil {
		return true, err
	}
	return true, nil
}

func finalize(msClient *v1.ComV1Client, k8sClient *kubernetes.Clientset, ms *types.ManagedSecret) {
	secretsClient := k8sClient.CoreV1().Secrets(ms.ObjectMeta.Namespace)
	err := secretsClient.Delete(ms.Status.Secret, &metav1.DeleteOptions{})
	if err != nil {
		statusErr, ok := err.(*errorsv1.StatusError)
		if !ok || statusErr.ErrStatus.Code != 404 {
			// If we can't delete the secret and it actually exists (or might), log our error
			// and retry finalization later
			ms.Status.Error = err.Error()
			log.Printf("Error finalizing ManagedSecret %s: %+v", ms.String(), err)
			_, err = msClient.ManagedSecrets(ms.Namespace).Update(ms)
			if err != nil {
				// Log _something_ here to make this pile of tragic errors less  baffling to the user
				// but don't handle it further because there isn’t a whole lot more we can do
				log.Printf("Error updating ManagedSecret %s with finalization error: %s", ms.String(), err)
			}
			return
		}
		err = removeFinalizer(msClient, ms)
		if err != nil {
			// Presumably we aren't going to successfully update the object status successfully since we _just_
			// failed to update it, so log the error and move on
			log.Printf("Error removing finalizer from ManagedSecret %s: %s", ms.String(), err)
			return
		}
		log.Printf("Successfully finalized ManagedSecret %s", ms.String())
	}
}

func removeFinalizer(msClient *v1.ComV1Client, ms *types.ManagedSecret) error {
	finalizers := ms.ObjectMeta.Finalizers
	newFinalizers := []string{}
	for _, finalizer := range finalizers {
		if finalizer != "managedsecrets.endclothing.com" {
			newFinalizers = append(newFinalizers, finalizer)
		}
	}
	ms.ObjectMeta.Finalizers = newFinalizers
	_, err := msClient.ManagedSecrets(ms.Namespace).Update(ms)
	if err != nil {
		return err
	}
	return nil
}

func isStale(ms *types.ManagedSecret) bool {
	return ms.Spec.Project != ms.Status.Project ||
		ms.Spec.Secret != ms.Status.Secret ||
		ms.Spec.Generation != ms.Status.Generation ||
		ms.Status.Error != ""
}

func createOrUpdateSecret(clientset *kubernetes.Clientset, secret *corev1.Secret) error {
	secretsClient := clientset.CoreV1().Secrets(secret.ObjectMeta.Namespace)
	_, err := secretsClient.Create(secret)
	if err != nil {
		statusErr, ok := err.(*errorsv1.StatusError)
		if ok {
			if statusErr.ErrStatus.Code == 409 {
				_, err = secretsClient.Update(secret)
			}
		}
	}
	return err
}
