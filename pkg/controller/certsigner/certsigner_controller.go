package certsigner

import (
	"context"
	"crypto/x509"
	"fmt"

	authorization "k8s.io/api/authorization/v1beta1"
	capi "k8s.io/api/certificates/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_certsigner")

// Tries to recognize CSRs that are specific to this use case
type csrRecognizer struct {
	recognize      func(csr *capi.CertificateSigningRequest, x509cr *x509.CertificateRequest) bool
	permission     authorization.ResourceAttributes
	successMessage string
}

func recognizers() []csrRecognizer {
	recognizers := []csrRecognizer{
		{
			recognize:      isCockroachServingCert,
			permission:     authorization.ResourceAttributes{Group: "certificates.k8s.io", Resource: "certificatesigningrequests", Verb: "create"},
			successMessage: "Auto approving CockroachDB certificate after SubjectAccessReview.",
		},
	}
	return recognizers
}

// Add creates a new CertSigner Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCertSigner{client: mgr.GetClient(), scheme: mgr.GetScheme(),  clientset: clientset.NewForConfigOrDie(mgr.GetConfig())}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("certsigner-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CertificateSigningRequest
	err = c.Watch(&source.Kind{Type: &capi.CertificateSigningRequest{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCertSigner{}

// ReconcileCertSigner reconciles a CertSigner object
type ReconcileCertSigner struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	// Helper client wrapper
	clientset clientset.Interface
}

// Reconcile reads that state of the cluster for a CertSigner object and makes changes based on the state read
// and what is in the CertSigner.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCertSigner) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CertSigner")

	// Fetch the CertificateSigningRequest instance
	csr := &capi.CertificateSigningRequest{}
	err := r.client.Get(context.TODO(), request.NamespacedName, csr)
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

	if len(csr.Status.Certificate) != 0 {
		log.Info("CSR already has a certificate, ignoring")
		return reconcile.Result{}, nil
	}

	if approved, denied := getCertApprovalCondition(&csr.Status); approved || denied {
		log.Info(fmt.Sprintf("CSR already has a approval status: %v", csr.Status))
		return reconcile.Result{}, nil
	}

	x509cr, err := parseCSR(csr)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("unable to parse csr %q: %v", csr.Name, err)
	}

	var tried []string

	for _, recognizer := range recognizers() {
		tried = append(tried, recognizer.permission.Resource)

		if !recognizer.recognize(csr, x509cr) {
			continue
		}

		approved, err := r.authorize(csr, recognizer.permission)
		if err != nil {
			log.Info(fmt.Sprintf("SubjectAccessReview failed:%s", err))
			return reconcile.Result{}, err
		}

		if approved {
			log.Info(fmt.Sprintf("approving csr %s with SANS: %s, IP Address:%s", csr.ObjectMeta.Name, x509cr.DNSNames, x509cr.IPAddresses))
			appendApprovalCondition(csr, recognizer.successMessage)
			_, err = r.clientset.CertificatesV1beta1().CertificateSigningRequests().UpdateApproval(csr)
			if err != nil {
				log.Info(fmt.Sprintf("error updating approval for csr: %v", err))
				return reconcile.Result{}, fmt.Errorf("error updating approval for csr: %v", err)
			}
		} else {
			log.Info(fmt.Sprintf("SubjectAccessReview not succesfull for CSR %s", request.NamespacedName))
			return reconcile.Result{}, fmt.Errorf("SubjectAccessReview not succesfull")
		}

		return reconcile.Result{}, nil

	}

	if len(tried) != 0 {
		log.Info(fmt.Sprintf("csr %s not recognized as Cockroach Node or Client csr, tried: %v", csr.Name, tried))
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

// Validate that the given node has authorization to actualy create CSRs
func (r *ReconcileCertSigner) authorize(csr *capi.CertificateSigningRequest, rattrs authorization.ResourceAttributes) (bool, error) {
	extra := make(map[string]authorization.ExtraValue)
	for k, v := range csr.Spec.Extra {
		extra[k] = authorization.ExtraValue(v)
	}

	sar := &authorization.SubjectAccessReview{
		Spec: authorization.SubjectAccessReviewSpec{
			User:               csr.Spec.Username,
			UID:                csr.Spec.UID,
			Groups:             csr.Spec.Groups,
			Extra:              extra,
			ResourceAttributes: &rattrs,
		},
	}
	sar, err := r.clientset.AuthorizationV1beta1().SubjectAccessReviews().Create(sar)
	if err != nil {
		return false, err
	}
	return sar.Status.Allowed, nil
}

func appendApprovalCondition(csr *capi.CertificateSigningRequest, message string) {
	csr.Status.Conditions = append(csr.Status.Conditions, capi.CertificateSigningRequestCondition{
		Type:    capi.CertificateApproved,
		Reason:  "AutoApproved",
		Message: message,
	})
}

