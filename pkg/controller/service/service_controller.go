package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"io"
	"strconv"
	"strings"

	pkcs12 "github.com/ykoer/serving-cert-keystore-operator/pkg/pkcs12"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_service")

const (
	// ServingCertSecretAnnotation Annotation used to inform the certificate generation service to
	// generate a cluster-signed certificate and populate the secret.
	servingCertSecretAnnotation = "service.alpha.openshift.io/serving-cert-secret-name"

	// ServingCertCreatePkcs12Annotation ...
	servingCertCreatePkcs12Annotation = "ykoer.github.com/serving-cert-create-pkcs12"

	tlsSecretCert               = "tls.crt"
	tlsSecretKey                = "tls.key"
	tlsPkcs12SecretFileName     = "tls.p12"
	tlsPkcs12SecretPasswordName = "tls-pkcs12-password"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Service Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("service-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileService{}

// ReconcileService reconciles a Service object
type ReconcileService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Service object and makes changes based on the state read
// and what is in the Service.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Service")

	// Fetch the Service instance
	instance := &corev1.Service{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	secretName := instance.ObjectMeta.Annotations[servingCertSecretAnnotation]
	createPkcs12, _ := strconv.ParseBool(instance.ObjectMeta.Annotations[servingCertCreatePkcs12Annotation])

	reqLogger.Info("Secret: " + secretName + ", Create Keystore: " + strconv.FormatBool(createPkcs12))

	if len(secretName) > 0 {
		request.Name = secretName
		secret := &corev1.Secret{}
		err := r.client.Get(context.TODO(), request.NamespacedName, secret)
		if err != nil {
			return reconcile.Result{}, err
		}

		// remove pkc12 keystore and password
		if !createPkcs12 && len(secret.Data[tlsPkcs12SecretFileName]) > 0 {
			r.removeServingCertSecretKeystore(secret)
			reqLogger.Info("Keystore removed!")
		}

		// create the pkcs12 keystore using the auto-generated crt and key
		if createPkcs12 && len(secret.Data[tlsPkcs12SecretFileName]) == 0 {
			r.createServingCertSecretKeystore(secret)
			reqLogger.Info("Keystore created!")
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileService) removeServingCertSecretKeystore(secret *corev1.Secret) (reconcile.Result, error) {
	delete(secret.Data, tlsPkcs12SecretFileName)
	delete(secret.Data, tlsPkcs12SecretPasswordName)
	err := r.client.Update(context.TODO(), secret)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Secret updated successfully
	return reconcile.Result{}, nil
}

func (r *ReconcileService) createServingCertSecretKeystore(secret *corev1.Secret) (reconcile.Result, error) {
	// create a random password
	password, err := getRandomPassword()
	if err != nil {
		return reconcile.Result{}, err
	}

	// create the pkcs12 keystore using the auto-generated crt and key
	pfxData, err := pkcs12.CreatePkcs12(secret.Data[tlsSecretCert], secret.Data[tlsSecretKey], password)

	secret.Data[tlsPkcs12SecretFileName] = pfxData
	secret.Data[tlsPkcs12SecretPasswordName] = []byte(password)

	err = r.client.Update(context.TODO(), secret)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Secret updated successfully
	return reconcile.Result{}, nil
}

func getRandomPassword() (string, error) {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(b), "="), nil
}
