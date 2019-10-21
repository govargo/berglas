package foo

import (
	"context"
	"encoding/base64"
	"github.com/GoogleCloudPlatform/berglas/pkg/berglas"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"strings"

	kwhhttp "github.com/slok/kubewebhook/pkg/http"
	kwhlog "github.com/slok/kubewebhook/pkg/log"
	kwhmutating "github.com/slok/kubewebhook/pkg/webhook/mutating"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	berglasDecrypt = "colopl.jp.berglas.decrypt"
)

var (
	secretData = map[string][]byte{}
)

// BerglasMutator is a mutator.
type BerglasMutator struct {
	logger kwhlog.Logger
}

// Mutate implements MutateFunc and provides the top-level entrypoint for object
// mutation.
func (m *BerglasMutator) Mutate(ctx context.Context, obj metav1.Object) (bool, error) {
	m.logger.Infof("calling mutate")

	secret, ok := obj.(*corev1.Secret)
	if !ok {
		m.logger.Errorf("error happens when cast object to secret")
		return true, nil
	}

	mutated := false

	for k, v := range secret.Data {
		d, didMutate, err := m.mutateSecretData(ctx, v)
		if err != nil {
			return true, err
		}
		if didMutate {
			mutated = true
			secretData[k] = d
		}
	}

	if !mutated {
		m.logger.Infof("there is no mutate")
		return false, nil
	}

	if mutated {
		if len(secret.Annotations) == 0 {
			secret.Annotations = make(map[string]string)
		}
		secret.Data = secretData
		secret.Annotations[berglasDecrypt] = "true"
		m.logger.Infof("The Secret resource %s is mutated", secret.GetObjectMeta().GetName())
	}

	return false, nil
}

func (m *BerglasMutator) mutateSecretData(ctx context.Context, data []byte) ([]byte, bool, error) {
	m.logger.Debugf("start mutating of secret data")
	decVal, isBerglasReference := m.hasBerglasReferences(data)
	if !isBerglasReference {
		m.logger.Infof("this secret resource does not have Barglas Reference.(i.e. berglas://${BUCKET_ID}/api-key)")
		return data, false, nil
	}

	bucket, object, err := parseRef(decVal)
	m.logger.Debugf("Target Bucket: %s", bucket)
	m.logger.Debugf("Target Object: %s", object)
	if err != nil {
		m.logger.Errorf("error parse berglas reference: %s", err)
		os.Exit(1)
	}

	acessRequest := berglas.AccessRequest{
		Bucket:     bucket,
		Object:     object,
		Generation: 0,
	}

	plainData, err := berglas.Access(ctx, &acessRequest)
	if err != nil {
		m.logger.Errorf("error decrypt secret by berglas: %s", err)
		return nil, false, err
	}
	m.logger.Debugf("berglas secret has been decrypted")
	plainByte := []byte(plainData)

	return plainByte, true, nil
}

func (m *BerglasMutator) hasBerglasReferences(data []byte) (string, bool) {
	decStr := string(data)
	if berglas.IsReference(decStr) {
		return decStr, true
	}
	return "", false
}

// webhookHandler is the http.Handler that responds to webhooks
func webhookHandler() http.Handler {
	logger := &kwhlog.Std{Debug: true}

	mutator := &BerglasMutator{logger: logger}

	mcfg := kwhmutating.WebhookConfig{
		Name: "berglasSecrets",
		Obj:  &corev1.Secret{},
	}

	// Create the wrapping webhook
	wh, err := kwhmutating.NewWebhook(mcfg, mutator, nil, nil, logger)
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

// parseRef parses a secret ref into a bucket, secret path, and any errors.
func parseRef(s string) (string, string, error) {
	s = strings.TrimPrefix(s, "gs://")
	s = strings.TrimPrefix(s, "berglas://")

	ss := strings.SplitN(s, "/", 2)
	if len(ss) < 2 {
		return "", "", errors.Errorf("secret does not match format gs://<bucket>/<secret> or the format berglas://<bucket>/<secret>: %s", s)
	}

	return ss[0], ss[1], nil
}

func byteToDecodeStr(b []byte) string {
	str := string(b)
	dec, _ := base64.StdEncoding.DecodeString(str)
	decStr := string(dec)

	return decStr
}