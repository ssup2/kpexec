package kpexec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Init package
func init() {
	rand.Seed(time.Now().UnixNano())
}

// Const
const (
	cnsPodDefaultTimeout = 30
	cnsPodLabelKey       = "kpexec.ssup2"
	cnsPodLabelValue     = "cnsenter"

	cnsContName             = "cnsenter"
	cnsContDefaultImg       = "ssup2/cnsenter:latest"
	cnsContDefaultToolsImg  = "ssup2/cnsenter-tools:latest"
	cnsContDefaultToolsRoot = "/croot"
	cnsContProcRemountExec  = "remount-proc-exec"

	contRuntimeContD  = "containerd"
	contRuntimeDocker = "docker"

	contDSocketVolume = "containerd-socket"
	contDSocketPath   = "/run/containerd/containerd.sock"

	contRootContdVolume  = "container-containerd-root"
	contRootContdPath    = "/run/containerd"
	contRootDockerVolume = "container-docker-root"
	contRootDockerPath   = "/var/lib/docker"

	kpexecExample = `
		# Get output from running 'date' command from pod mypod, using the first container by default
		kpexec mypod -- date

		# Get output from running 'date' command in date-container from pod mypod and namespace mynamespace 
		kpexec -n mynamespace mypod -c date-container -- date

		# Switch to raw terminal mode, sends stdin to 'bash' in bash-container from pod mypod
		# and sends stdout/stderr from 'bash' back to the client
		kpexec -it mypod -c bash-container -- bash 

		# Enable 'tools' mode
		kpexec -it -T mypod -c bash-container -- bash 

		# Set cnsenter pod's image
		kpexec -it -T --cnsenter-img=ssup2/my-cnsenter-tools:latest mypod -c golang-container -- bash

		# Run cnsenter pod garbage collector
		kpexec --cnsenter-gc
		`
)

// Cmd
func New() *cobra.Command {
	// Get cobra cmd
	options := &Options{}
	cmd := &cobra.Command{
		Use:                   "kpexec [-n NAMESPACE] POD [-c CONTAINER] -- COMMAND [args...]",
		DisableFlagsInUseLine: true,
		Short:                 "Execute a command with privilige in a container.",
		Long:                  "Execute a command with privilige in a container.",
		Example:               kpexecExample,
		Run: func(cmd *cobra.Command, args []string) {
			if len(options.completion) != 0 {
				if err := options.Complete(cmd, args); err != nil {
					fmt.Printf("Failed to get bash/zsh completion : %+v\n", err)
					os.Exit(1)
				}
			} else if options.cnsPodGC {
				if err := options.GarbageCollect(); err != nil {
					fmt.Printf("Failed to run cnsenter pod's garbage collector : %+v\n", err)
					os.Exit(1)
				}
			} else {
				if err := options.Run(args, cmd.ArgsLenAtDash()); err != nil {
					fmt.Printf("Failed to run kpexec err : %+v\n", err)
					os.Exit(1)
				}
			}
		},
		BashCompletionFunction: bashCompletionFunc,
	}

	// Set flags
	cmd.Flags().StringVarP(&options.tPodNs, "namespace", "n", "", "If present, the namespace scope for this CLI request")
	cmd.Flags().StringVarP(&options.tContName, "container", "c", "", "Container name. If omitted, the first container in the pod will be chosen")
	cmd.Flags().BoolVarP(&options.stdin, "stdin", "i", false, "Pass stdin to the container")
	cmd.Flags().BoolVarP(&options.tty, "tty", "t", false, "Stdin is a TTY")
	cmd.Flags().BoolVarP(&options.tools, "tools", "T", false, "Use tools mode")

	cmd.Flags().StringVar(&options.cnsPodNamespace, "cnsenter-ns", "", "Set cnsenter pod's namespace (default target pod's namespace)")
	cmd.Flags().StringVar(&options.cnsPodImage, "cnsenter-img", "", "Set cnsenter pod's img (default mode ssup2/cnsenter:latest / tools mode ssup2/cnsenter-tools:latest)")
	cmd.Flags().Int32Var(&options.cnsPodTimeout, "cnsenter-to", cnsPodDefaultTimeout, "Set cnsenter pod's creation timeout")
	cmd.Flags().BoolVar(&options.cnsPodGC, "cnsenter-gc", false, "Run cnsenter pod garbage collector")

	cmd.Flags().StringVar(&options.kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "Absolute path to the kubeconfig file")
	cmd.Flags().StringVar(&options.completion, "completion", "", "Output shell completion code for the specified shell (bash or zsh)")

	// Set bash completion flags
	for name, completion := range bashCompletionFlags {
		cmd.Flag(name).Annotations = map[string][]string{}
		cmd.Flag(name).Annotations[cobra.BashCompCustom] = append(
			cmd.Flag(name).Annotations[cobra.BashCompCustom],
			completion,
		)
	}

	return cmd
}

// PexecOptions
type Options struct {
	tPodNs    string
	tContName string
	tty       bool
	stdin     bool
	tools     bool

	cnsPodNamespace string
	cnsPodImage     string
	cnsPodTimeout   int32
	cnsPodGC        bool

	kubeconfig string
	completion string
}

func (o *Options) Complete(cmd *cobra.Command, args []string) error {
	if o.completion == "bash" {
		// Print out bash completion
		if err := cmd.GenBashCompletion(os.Stdout); err != nil {
			return fmt.Errorf("failed to generate bash completion : %+v", err)
		}
		return nil
	} else if o.completion == "zsh" {
		// Get bash completion
		compBuf := new(bytes.Buffer)
		if err := cmd.GenBashCompletion(compBuf); err != nil {
			return fmt.Errorf("failed to generate zsh completion : %+v", err)
		}

		// Merge zsh head, bash completion, zsh tail and print out
		mergedBytes := []byte(zshHead)
		mergedBytes = append(mergedBytes, compBuf.Bytes()...)
		mergedBytes = append(mergedBytes, []byte(zshTail)...)
		if _, err := os.Stdout.Write(mergedBytes); err != nil {
			return fmt.Errorf("failed to generate zsh completion : %+v", err)
		}
		return nil
	}
	return fmt.Errorf("%s is not supported shell", o.completion)
}

func (o *Options) GarbageCollect() error {
	// Init k8s client
	config, err := clientcmd.BuildConfigFromFlags("", o.kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to set kubeconfig : %+v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to set clientset : %+v", err)
	}

	// Get target pod's info
	cnsPods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		LabelSelector: cnsPodLabelKey + "=" + cnsPodLabelValue,
	})
	if err != nil {
		return fmt.Errorf("failed to all cnsenter pod's list : %+v", err)
	}

	for _, cnsPod := range cnsPods.Items {
		// Check cnsenter pod is running status
		if cnsPod.Status.Phase == corev1.PodRunning {
			continue
		}

		// Delete cnsenter pod
		if err := clientset.CoreV1().Pods(cnsPod.Namespace).Delete(context.TODO(), cnsPod.Name, metav1.DeleteOptions{}); err != nil {
			fmt.Printf("Failed to delete to cnsenter pod : %+v\n", err)
		}
	}
	return nil
}

func (o *Options) Run(args []string, argsLenAtDash int) error {
	// Check inputs
	// Check pod name by using double dash
	if argsLenAtDash == -1 {
		return fmt.Errorf("no double dash")
	} else if argsLenAtDash == 0 {
		return fmt.Errorf("no target pod name")
	} else if argsLenAtDash >= 2 {
		return fmt.Errorf("wrong pod name")
	}
	// Chech commands
	if len(args) <= 1 {
		return fmt.Errorf("no commands")
	}

	tPodName := args[argsLenAtDash-1]
	tPodCmd := args[argsLenAtDash:]

	// Init k8s client
	config, err := clientcmd.BuildConfigFromFlags("", o.kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to set kubeconfig : %+v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to set clientset : %+v", err)
	}

	// If not set target pod's names, Get default namespace from kubeconfig
	if o.tPodNs == "" {
		cmdConfig, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
		if err != nil {
			return fmt.Errorf("failed to get default namespace : %+v", err)
		}
		ns := cmdConfig.Contexts[cmdConfig.CurrentContext].Namespace
		if ns == "" {
			o.tPodNs = "default"
		} else {
			o.tPodNs = ns
		}
	}

	// Get target pod's info
	tPod, err := clientset.CoreV1().Pods(o.tPodNs).Get(context.TODO(), tPodName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get target pod's info : %+v", err)
	}
	tPodNodeName := tPod.Spec.NodeName
	if o.tContName == "" {
		// Get first container name
		o.tContName = tPod.Spec.Containers[0].Name
		fmt.Printf("Defaulting container name to %s.\n", o.tContName)
	}

	// Get target container's info
	tContRuntime, tContID, err := getContainerRuntimeID(tPod, o.tContName)
	if err != nil {
		return fmt.Errorf("failed to get target container's info : %+v", err)
	}

	// Create and set defer to delete cnsenter pod
	// Config cnsenter pod
	cnsPodName := fmt.Sprintf("cnsenter-%s", getRandomString(10))
	cnsDSocketVolumeType := corev1.HostPathSocket
	cnsContRootVolumeType := corev1.HostPathDirectory
	cnsPrivileged := true

	cnsPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: cnsPodName,
			Labels: map[string]string{
				cnsPodLabelKey: cnsPodLabelValue,
			},
		},
		Spec: corev1.PodSpec{
			NodeName: tPodNodeName,
			Containers: []corev1.Container{
				{
					Name:  cnsContName,
					Image: cnsContDefaultImg,
					Stdin: o.stdin,
					TTY:   o.tty,
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      contDSocketVolume,
							MountPath: contDSocketPath,
						},
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: &cnsPrivileged,
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: contDSocketVolume,
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: contDSocketPath,
							Type: &cnsDSocketVolumeType,
						},
					},
				},
			},
			Tolerations: []corev1.Toleration{
				{
					Operator: corev1.TolerationOpExists,
				},
			},
			HostPID:       true,
			RestartPolicy: "Never",
		},
	}

	if o.tools {
		// For tools mode
		// Use tools image
		cnsPod.Spec.Containers[0].Image = cnsContDefaultToolsImg

		// Set command
		// Do not enter mount namespace
		// Create new mount namespace and remount procfs
		cnsPodCmd := []string{"cnsenter", "--pid", "--net", "--ipc", "--uts",
			"--runtime=" + tContRuntime, "--container=" + tContID,
			"--root-symlink", cnsContDefaultToolsRoot, "--wd", "--wd-base", cnsContDefaultToolsRoot,
			"--env", "TERM=xterm", "--", "unshare", "--mount", cnsContProcRemountExec}
		cnsPodCmd = append(cnsPodCmd, tPodCmd...)
		cnsPod.Spec.Containers[0].Command = cnsPodCmd

		// Copy DNS settings from target pods
		cnsPod.Spec.DNSPolicy = tPod.Spec.DeepCopy().DNSPolicy
		cnsPod.Spec.DNSConfig = tPod.Spec.DeepCopy().DNSConfig

		// Set volume to access
		if tContRuntime == contRuntimeContD {
			cnsPod.Spec.Volumes = append(cnsPod.Spec.Volumes,
				corev1.Volume{
					Name: contRootContdVolume,
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: contRootContdPath,
							Type: &cnsContRootVolumeType,
						},
					},
				})

			cnsPod.Spec.Containers[0].VolumeMounts = append(cnsPod.Spec.Containers[0].VolumeMounts,
				corev1.VolumeMount{
					Name:      contRootContdVolume,
					MountPath: contRootContdPath,
				})
		} else if tContRuntime == contRuntimeDocker {
			cnsPod.Spec.Volumes = append(cnsPod.Spec.Volumes,
				corev1.Volume{
					Name: contRootDockerVolume,
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: contRootDockerPath,
							Type: &cnsContRootVolumeType,
						},
					},
				})

			cnsPod.Spec.Containers[0].VolumeMounts = append(cnsPod.Spec.Containers[0].VolumeMounts,
				corev1.VolumeMount{
					Name:      contRootDockerVolume,
					MountPath: contRootDockerPath,
				})
		} else {
			return fmt.Errorf("%s is not supported container runtime", tContRuntime)
		}
	} else {
		// For default mode
		// Use default image
		cnsPod.Spec.Containers[0].Image = cnsContDefaultImg

		// Set command
		cnsPodCmd := []string{"cnsenter", "--mount", "--pid", "--net", "--ipc", "--uts",
			"--runtime=" + tContRuntime, "--container=" + tContID, "--wd", "--"}
		cnsPodCmd = append(cnsPodCmd, tPodCmd...)
		cnsPod.Spec.Containers[0].Command = cnsPodCmd
	}

	// Set cnsenter pod's namespace and image
	if o.cnsPodNamespace == "" {
		o.cnsPodNamespace = o.tPodNs
	}
	if o.cnsPodImage != "" {
		cnsPod.Spec.Containers[0].Image = o.cnsPodImage
	}

	// Create a cnsenter pod
	_, err = clientset.CoreV1().Pods(o.cnsPodNamespace).Create(context.TODO(), cnsPod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create cnsetner pod : %+v", err)
	}
	defer func() {
		// Delete cnsenter pod
		if err := clientset.CoreV1().Pods(o.cnsPodNamespace).Delete(context.TODO(), cnsPodName, metav1.DeleteOptions{}); err != nil {
			fmt.Printf("Failed to delete to cnsenter pod : %+v\n", err)
		}
	}()

	// Set signal handler to delete cnsenterpod
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigs
		fmt.Printf("Recived signal %s\n", sig)

		// Delete cnsenter pod and exit
		if err := clientset.CoreV1().Pods(o.cnsPodNamespace).Delete(context.TODO(), cnsPodName, metav1.DeleteOptions{}); err != nil {
			fmt.Printf("Failed to delete to cnsenter pod : %+v\n", err)
		}
		os.Exit(1)
	}()

	// Wait cnsenter container to run
	// Watch cnsenter pod
	cnsPodWatch, err := clientset.CoreV1().Pods(o.cnsPodNamespace).
		Watch(context.TODO(), metav1.ListOptions{Watch: true, FieldSelector: "metadata.name=" + cnsPodName})
	if err != nil {
		return fmt.Errorf("failed to wait running cnsenter pod : %+v", err)
	}
	// Set wait timeout
	cnsPodTimer := time.NewTimer(time.Duration(o.cnsPodTimeout) * time.Second)
	go func() {
		<-cnsPodTimer.C
		fmt.Printf("Failed to wait cnsenter pod\n")

		// Delete cnsenter pod and exit
		if err := clientset.CoreV1().Pods(o.cnsPodNamespace).Delete(context.TODO(), cnsPodName, metav1.DeleteOptions{}); err != nil {
			fmt.Printf("Failed to delete to cnsenter pod : %+v\n", err)
		}
		os.Exit(1)
	}()
	// Wait and check pod's status
	for cnsPodEvent := range cnsPodWatch.ResultChan() {
		tPod, _ = cnsPodEvent.Object.(*corev1.Pod)
		if tPod.Status.Phase == corev1.PodRunning || tPod.Status.Phase == corev1.PodSucceeded || tPod.Status.Phase == corev1.PodFailed {
			break
		}
	}
	// Stop timer and watch
	cnsPodTimer.Stop()
	cnsPodWatch.Stop()

	// Attach cnsenter pod
	if (o.tty || o.stdin) && tPod.Status.Phase == corev1.PodRunning {
		if err := attachPod(o.cnsPodNamespace, cnsPodName, "cnsenter", o.tty, o.stdin); err == nil {
			return nil
		}

		// Check cnsenter pod terminated through error
		if !strings.Contains(err.Error(), "completed pod") {
			return fmt.Errorf("failed to attach to cnsenter pod : %+v", err)
		}
		// If cnsenter pod is terminated, get it's log
	}

	// Get cnsenter pod's logs
	cnsLogReq := clientset.CoreV1().Pods(o.cnsPodNamespace).GetLogs(cnsPodName, &corev1.PodLogOptions{Follow: true})
	cnsLog, err := cnsLogReq.Stream(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get cnsenter pod's log stream : %+v", err)
	}
	defer cnsLog.Close()

	// Print cnsenter pod's logs
	for {
		n, err := io.Copy(os.Stdout, cnsLog)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to get cnsenter pod's log : %+v", err)
		}
	}

	return nil
}

// Helpers
func getContainerRuntimeID(pod *corev1.Pod, containerName string) (string, string, error) {
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containerName {
			u, err := url.Parse(status.ContainerID)
			if err != nil {
				return "", "", fmt.Errorf("parse container ID error")
			}
			return u.Scheme, u.Host, nil
		}
	}

	return "", "", fmt.Errorf("no container runtime, ID info")
}

func getRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func attachPod(podNs, podName, contName string, tty, stdin bool) error {
	// Set kubectl attach args
	args := []string{"attach", "--namespace", podNs, podName, "--container", contName}
	if tty {
		args = append(args, "--tty")
	}
	if stdin {
		args = append(args, "--stdin")
	}

	// Run kubectl attach
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
