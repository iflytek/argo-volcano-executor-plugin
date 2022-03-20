package controller

import (
	"fmt"
	"testing"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

func TestVcJobObjectMeta(T *testing.T) {
	vcjob := &v1alpha1.Job{}
	lk := map[string]string{
		"a": "b",
	}
	vcjob.ObjectMeta.Labels = lk
	fmt.Println(vcjob)
}
