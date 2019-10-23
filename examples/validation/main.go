package foo

import (
	"context"
	"fmt"
	"github.com/GoogleCloudPlatform/berglas/pkg/berglas"
	"net/http"
	"os"

	kwhhttp "github.com/slok/kubewebhook/pkg/http"
	kwhlog "github.com/slok/kubewebhook/pkg/log"
	kwhvalidating "github.com/slok/kubewebhook/pkg/webhook/validating"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BerglasValidator struct {
	logger kwhlog.Logger
}

// Validate implements ValidateFunc and provides the top-level entrypoint for object
// validation.
func (v *BerglasValidator) Validate(ctx context.Context, obj metav1.Object) (bool, kwhvalidating.ValidatorResult, error) {
	v.logger.Infof("calling validate")

	secret, ok := obj.(*corev1.Secret)
	if !ok {
		v.logger.Errorf("error happens when cast object to secret")
		return false, kwhvalidating.ValidatorResult{}, fmt.Errorf("not an ingress")
	}

	for key, val := range secret.Data {
		validateError := v.validateSecretData(ctx, val)
		if validateError {
			res := kwhvalidating.ValidatorResult{
				Valid:   false,
				Message: fmt.Sprintf("the secret data: %s invalid berglas reference. the data cannot be decrypted.", key),
			}
			return false, res, nil
		}
	}

	v.logger.Infof("secret %s is valid", secret.Name)
	res := kwhvalidating.ValidatorResult{
		Valid:   true,
		Message: "all secretData are valid",
	}
	return false, res, nil
}

func (v *BerglasValidator) validateSecretData(ctx context.Context, data []byte) (bool) {
	v.logger.Debugf("start validating of secret data")
	isBerglasReference := v.hasBerglasReferences(data)
	if !isBerglasReference {
		return false
	}

	v.logger.Infof("this secret resource has Berglas Reference(i.e. berglas://${BUCKET_ID}/api-key). the Berglas data must be decrypted.")
	return true
}

func (v *BerglasValidator) hasBerglasReferences(data []byte) (bool) {
	secretVal := string(data)
	if berglas.IsReference(secretVal) {
		return true
	}
	return false
}

// webhookHandler is the http.Handler that responds to webhooks
func webhookHandler() http.Handler {
	logger := &kwhlog.Std{Debug: true}

	vl := &BerglasValidator{logger: logger}

	vcfg := kwhvalidating.WebhookConfig{
		Name: "berglasSecrets",
		Obj:  &corev1.Secret{},
	}

	// Create the wrapping webhook
	wh, err := kwhvalidating.NewWebhook(vcfg, vl, nil, nil, logger)
	if err != nil {
		logger.Errorf("error creating webhook: %s", err)
		os.Exit(1)
	}

	// Get the handler for our webhook.
	whhandler, err := kwhhttp.HandlerFor(wh)
	if err != nil {
		logger.Errorf("error creating webhook handler: %s", err)
		os.Exit(1)
	}
	return whhandler
}

// F is the exported webhook for the function to bind.
var F = webhookHandler().ServeHTTP