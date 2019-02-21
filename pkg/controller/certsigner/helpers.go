package certsigner

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"reflect"
	"strings"

	capi "k8s.io/api/certificates/v1beta1"
)

func getCertApprovalCondition(status *capi.CertificateSigningRequestStatus) (approved bool, denied bool) {
	for _, c := range status.Conditions {
		if c.Type == capi.CertificateApproved {
			approved = true
		}
		if c.Type == capi.CertificateDenied {
			denied = true
		}
	}
	return
}

func isApproved(csr *capi.CertificateSigningRequest) bool {
	approved, denied := getCertApprovalCondition(&csr.Status)
	return approved && !denied
}

// parseCSR extracts the CSR from the API object and decodes it.
func parseCSR(obj *capi.CertificateSigningRequest) (*x509.CertificateRequest, error) {
	// extract PEM from request object
	pemBytes := obj.Spec.Request
	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return nil, errors.New("PEM block type must be CERTIFICATE REQUEST")
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, err
	}
	return csr, nil
}

func hasExactUsages(csr *capi.CertificateSigningRequest, usages []capi.KeyUsage) bool {
	if len(usages) != len(csr.Spec.Usages) {
		return false
	}

	usageMap := map[capi.KeyUsage]struct{}{}
	for _, u := range usages {
		usageMap[u] = struct{}{}
	}

	for _, u := range csr.Spec.Usages {
		if _, ok := usageMap[u]; !ok {
			return false
		}
	}

	return true
}

var nodeUsages = []capi.KeyUsage{
	capi.UsageKeyEncipherment,
	capi.UsageDigitalSignature,
	capi.UsageServerAuth,
	capi.UsageClientAuth,
}

var clientUsages = []capi.KeyUsage{
	capi.UsageKeyEncipherment,
	capi.UsageDigitalSignature,
	capi.UsageClientAuth,
}

func isCockroachServingCert(csr *capi.CertificateSigningRequest, x509cr *x509.CertificateRequest) bool {
	log.Info(fmt.Sprintf("Introspecting cert: %+v", x509cr.Subject))
	if !reflect.DeepEqual([]string{"Cockroach"}, x509cr.Subject.Organization) {
		log.Info(fmt.Sprintf("Org does not match: %s", x509cr.Subject.Organization))
		return false
	}
	if strings.HasPrefix(x509cr.Subject.CommonName, "node") {
		if !hasExactUsages(csr, nodeUsages) {
			log.Info(fmt.Sprintf("Usage does not match: %s", csr.Spec.Usages))
			return false
		} else {
			return true
		}
	}
	if strings.HasPrefix(x509cr.Subject.CommonName, "root") {
		if !hasExactUsages(csr, clientUsages) {
			log.Info(fmt.Sprintf("Usage does not match: %s", csr.Spec.Usages))
			return false
		} else {
			return true
		}
	}
	return false
}
