package foo

import (
	"context"
	kwhlog "github.com/slok/kubewebhook/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)


func Test_Mutate_Berglas(t *testing.T) {
	t.Run("secret resource has berglas reference", func(t *testing.T) {

		expect := false
		testData := map[string]string{
			"API_KEY": "berglas://verification-iso-berglas-secret/api-key",
			"TLS_KEY": "berglas://verification-iso-berglas-secret/tls-key",
		}

		ctx := context.Background()
		logger := &kwhlog.Std{Debug: true}
		mutator := &BerglasMutator{logger: logger}
		secretData := map[string][]byte{}

		for k, _ := range testData {
			secretData[k] = []byte(testData[k])
		}

		secret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testsecret",
				Namespace: "default",
			},
			Data:       secretData,
			StringData: nil,
			Type:       "",
		}

		isMutated, err := mutator.Mutate(ctx, &secret)
		if err != nil {
			t.Fatal(err)
		}
		if isMutated != expect {
			t.Fail()
		} else {
			println("Test Sucessed")
		}
	})
}

func Test_Mutate_BerglasLess(t *testing.T) {
	t.Run("secret resource does not have berglas reference", func(t *testing.T) {

		expect := false
		testData := map[string]string{
			"API_KEY": "abcd1234",
			"TLS_KEY": "efgh5678",
		}

		ctx := context.Background()
		logger := &kwhlog.Std{Debug: true}
		mutator := &BerglasMutator{logger: logger}
		secretData := map[string][]byte{}

		for k, v := range testData {
			secretData[k] = []byte(v)
		}

		secret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testsecret",
				Namespace: "default",
			},
			Data:       secretData,
			StringData: nil,
			Type:       "",
		}

		isMutated, err := mutator.Mutate(ctx, &secret)
		if err != nil {
			t.Fatal(err)
		}
		if isMutated != expect {
			t.Fail()
		} else {
			println("Test Sucessed")
		}
	})
}