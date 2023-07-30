package main

import (
	"context"
	"fmt"
    batchv1 "k8s.io/api/batch/v1"
    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    kubernetes "k8s.io/client-go/kubernetes"
    clientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/retry"
)

func main() {
	fmt.Println("Start backup creation")

	kubeconfig := "/var/run/secrets/kubernetes.io/serviceaccount/kubeconfig"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set{"backup": "true"}.String(),
	}

	pvList, err := clientset.CoreV1().PersistentVolumes().List(context.TODO(), listOptions)
	if err != nil {
		panic(err.Error())
	}

	for _, pv := range pvList.Items {
		fmt.Printf("Found PV with annotation backup:true: %s\n", pv.Name)
		createJob(clientset, pv.Name)
	}
}

func createJob(clientset *kubernetes.Clientset, pvName string) {
	podName := "job-pod"+pvName
	containerName := "job-container"
	jobName := "backup-job-"+pvName

	podSpec := v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:  containerName,
				Image: "your-backup-image",
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "backup-volume",
						MountPath: "/app",
					},
				},
			},
		},
		RestartPolicy: v1.RestartPolicyNever,
		Volumes: []v1.Volume{
			{
				Name: "backup-volume",
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvName,
					},
				},
			},
		},
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   podName,
					Labels: map[string]string{"job": jobName},
				},
				Spec: podSpec,
			},
		},
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, err := clientset.BatchV1().Jobs("pv-backup").Create(context.TODO(), job, metav1.CreateOptions{})
		return err
	})
	if retryErr != nil {
		panic(fmt.Errorf("error creating job: %v", retryErr))
	}
}