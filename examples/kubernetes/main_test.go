package foo

import (
	"context"
	kwhlog "github.com/slok/kubewebhook/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)


func Test_Mutate(t *testing.T) {
	t.Run("text", func(t *testing.T) {

		expect := false
		ctx := context.Background()
		logger := &kwhlog.Std{Debug: true}
		mutator := &BerglasMutator{logger: logger}
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testpod",
				Namespace: "default",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "nginx",
						Image: "nginx:latest",
						Command: []string{"/bin/echo"},
						Env: []corev1.EnvVar{
							{
								Name:  "API_KEY",
								Value: "berglas://verification-iso-berglas-secret/api-key",
							},
						},
					},
				},
			},
		}

		r, err := mutator.Mutate(ctx, &pod)
		if err != nil {
			t.Fatal(err)
		}
		if r != expect {
			t.Fail()
		} else {
			println("Test Sucessed")
		}
	})
}