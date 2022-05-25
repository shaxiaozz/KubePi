package podbase

import (
	"bytes"
	"context"
	"fmt"
	"github.com/KubeOperator/kubepi/pkg/util/podexec"
	"io"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const KotoolsPath = "/kotools"

type PodBase struct {
	Namespace  string
	PodName    string
	Container  string
	K8sClient  *kubernetes.Clientset
	RestClient *rest.Config
}

func (p *PodBase) NewPodExec() *podexec.PodExec {
	return podexec.NewPodExec(p.Namespace, p.PodName, p.Container, p.RestClient, p.K8sClient)
}

func (p *PodBase) PodInfo() (*coreV1.Pod, error) {
	pod, err := p.K8sClient.CoreV1().Pods(p.Namespace).
		Get(context.TODO(), p.PodName, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if pod.Status.Phase == coreV1.PodSucceeded || pod.Status.Phase == coreV1.PodFailed {
		return nil, fmt.Errorf("cannot exec into a container in a completed pod; current phase is %s", pod.Status.Phase)
	}
	return pod, nil
}

func (p *PodBase) OsAndArch(nodeName string) (osType string, arch string) {
	// get pod system arch and type
	node, err := p.K8sClient.CoreV1().Nodes().
		Get(context.TODO(), nodeName, metaV1.GetOptions{})
	if err == nil {
		var ok bool
		osType, ok = node.Labels["beta.kubernetes.io/os"]
		if !ok {
			osType, ok = node.Labels["kubernetes.io/os"]
			if !ok {
				osType = "linux"
			}
		}
		arch, ok = node.Labels["beta.kubernetes.io/arch"]
		if !ok {
			arch, ok = node.Labels["kubernetes.io/arch"]
			if !ok {
				arch = "amd64"
			}
		}
	}
	return
}

func (p *PodBase) Exec(stdin io.Reader, command ...string) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	exec := p.NewPodExec()
	exec.Command = command
	exec.Tty = false
	exec.Stdin = stdin
	exec.Stdout = &stdout
	exec.Stderr = &stderr
	err := exec.Exec(podexec.Exec)
	if err != nil {
		if len(stderr.Bytes()) != 0 {
			return nil, fmt.Errorf("%s %s", err.Error(), stderr.String())
		}
		return nil, err
	}
	if len(stderr.Bytes()) != 0 {
		return nil, fmt.Errorf(stderr.String())
	}
	return stdout.Bytes(), nil
}

func (p *PodBase) InstallKOTools() error {
	pod, err := p.PodInfo()
	if err != nil {
		return err
	}
	osType, arch := p.OsAndArch(pod.Spec.NodeName)
	kfToolsPath := fmt.Sprintf(KotoolsPath+"/kotools_%s_%s", osType, arch)
	if osType == "windows" {
		kfToolsPath += ".exe"
	}
	exec := p.NewPodExec()
	err = exec.CopyToPod(kfToolsPath, "/kotools")
	if err != nil {
		return err
	}
	if osType != "windows" {
		chmodCmd := []string{"chmod", "+x", "/kotools"}
		exec.Command = chmodCmd
		var stderr bytes.Buffer
		exec.Stderr = &stderr
		err = exec.Exec(podexec.Exec)
		if err != nil {
			return fmt.Errorf(err.Error(), stderr)
		}
	}
	return nil
}
