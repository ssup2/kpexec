package kpexec

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"k8s.io/kubectl/pkg/cmd/exec"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

const (
	kpexecUsageStr        = "expected 'kpexec [-n namespace] (POD | TYPE/NAME) COMMAND [ARG1] [ARG2] ... [ARGN]'.\nPOD or TYPE/NAME and COMMAND are required arguments for the exec command"
	defaultPodExecTimeout = 60 * time.Second

	hostDirectory        = "/host"
	containerdSocketFile = "/run/containerd/containerd.sock"

	bashCompletionFunc = `# call kubectl get $1,
__kubectl_debug_out()
{
    local cmd="$1"
    __kubectl_debug "${FUNCNAME[1]}: get completion by ${cmd}"
    eval "${cmd} 2>/dev/null"
}

__kubectl_override_flag_list=(--kubeconfig --cluster --user --context --namespace --server -n -s)
__kubectl_override_flags()
{
    local ${__kubectl_override_flag_list[*]##*-} two_word_of of var
    for w in "${words[@]}"; do
        if [ -n "${two_word_of}" ]; then
            eval "${two_word_of##*-}=\"${two_word_of}=\${w}\""
            two_word_of=
            continue
        fi
        for of in "${__kubectl_override_flag_list[@]}"; do
            case "${w}" in
                ${of}=*)
                    eval "${of##*-}=\"${w}\""
                    ;;
                ${of})
                    two_word_of="${of}"
                    ;;
            esac
        done
    done
    for var in "${__kubectl_override_flag_list[@]##*-}"; do
        if eval "test -n \"\$${var}\""; then
            eval "echo -n \${${var}}' '"
        fi
    done
}

# $1 is the name of resource (required)
# $2 is template string for kubectl get (optional)
__kubectl_parse_get()
{
    local template
    template="${2:-"{{ range .items  }}{{ .metadata.name }} {{ end }}"}"
    local kubectl_out
    if kubectl_out=$(__kubectl_debug_out "kubectl get $(__kubectl_override_flags) -o template --template=\"${template}\" \"$1\""); then
        COMPREPLY+=( $( compgen -W "${kubectl_out[*]}" -- "$cur" ) )
    fi
}

__kubectl_get_resource_namespace()
{
    __kubectl_parse_get "namespace"
}

__kubectl_get_resource_pod()
{
    __kubectl_parse_get "pod"
}

# $1 is the name of the pod we want to get the list of containers inside
__kubectl_get_containers()
{
    local template
    template="{{ range .spec.initContainers }}{{ .name }} {{end}}{{ range .spec.containers  }}{{ .name }} {{ end }}"
    __kubectl_debug "${FUNCNAME} nouns are ${nouns[*]}"

    local len="${#nouns[@]}"
    if [[ ${len} -ne 1 ]]; then
        return
    fi
    local last=${nouns[${len} -1]}
    local kubectl_out
    if kubectl_out=$(__kubectl_debug_out "kubectl get $(__kubectl_override_flags) -o template --template=\"${template}\" pods \"${last}\""); then
        COMPREPLY=( $( compgen -W "${kubectl_out[*]}" -- "$cur" ) )
    fi
}

__kpexec_custom_func() {
	case ${last_command} in
        kpexec)
			__kubectl_get_resource_pod
            return
            ;;
        *)
            ;;
    esac
}
`

	zshHead = `#compdef kubectl

# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

__kubectl_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand

	source "$@"
}

__kubectl_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift

		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__kubectl_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}

__kubectl_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?

	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}

__kubectl_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}

__kubectl_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}

__kubectl_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}

__kubectl_filedir() {
	local RET OLD_IFS w qw

	__kubectl_debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi

	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"

	IFS="," __kubectl_debug "RET=${RET[@]} len=${#RET[@]}"

	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__kubectl_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}

__kubectl_quote() {
    if [[ $1 == \'* || $1 == \"* ]]; then
        # Leave out first character
        printf %q "${1:1}"
    else
	printf %q "$1"
    fi
}

autoload -U +X bashcompinit && bashcompinit

# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi

__kubectl_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__kubectl_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__kubectl_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__kubectl_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__kubectl_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__kubectl_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/builtin declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__kubectl_type/g" \
	<<'BASH_COMPLETION_EOF'
`

	zshTail = `
BASH_COMPLETION_EOF
}

__kubectl_bash_source <(__kubectl_convert_bash_to_zsh)
_complete kubectl 2>/dev/null
`
)

var (
	kpexecExample = templates.Examples(i18n.T(`
		# Get output from running 'date' command from pod mypod, using the first container by default
		kpexec mypod date

		# Get output from running 'date' command in ruby-container from pod mypod and namespace ruby
		kpexec -n ruby mypod -c ruby-container date

		# Switch to raw terminal mode, sends stdin to 'bash' in ruby-container from pod mypod
		# and sends stdout/stderr from 'bash' back to the client
		kpexec mypod -c ruby-container -i -t -- bash -il

		# List contents of /usr from the first container of pod mypod and sort by modification time.
		# If the command you want to execute in the pod has any flags in common (e.g. -i),
		# you must use two dashes (--) to separate your command's flags/arguments.
		# Also note, do not surround your command and its flags/arguments with quotes
		# unless that is how you would execute it normally (i.e., do ls -t /usr, not "ls -t /usr").
		kpexec mypod -i -t -- ls -t /usr

		# Get output from running 'date' command from the first pod of the deployment mydeployment, using the first container by default
		kpexec deploy/mydeployment date

		# Get output from running 'date' command from the first pod of the service myservice, using the first container by default
		kpexec svc/myservice date
		`))

	bashCompletionFlags = map[string]string{
		"namespace": "__kubectl_get_resource_namespace",
		"container": "__kubectl_get_containers",
	}
)

// Cmd
func NewCmdKpexec(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	// Get cobra cmd
	options := &PexecOptions{
		ExecOptions: exec.ExecOptions{
			StreamOptions: exec.StreamOptions{
				IOStreams: streams,
			},
			Executor: &exec.DefaultRemoteExecutor{},
		},
	}
	cmd := &cobra.Command{
		Use:                   "kpexec [-n namespace] (POD | TYPE/NAME) [-c CONTAINER] -- COMMAND [args...]",
		DisableFlagsInUseLine: true,
		Short:                 "Execute a command with privilige in a container.",
		Long:                  "Execute a command with privilige in a container.",
		Example:               kpexecExample,
		Run: func(cmd *cobra.Command, args []string) {
			argsLenAtDash := cmd.ArgsLenAtDash()
			cmdutil.CheckErr(options.Complete(f, cmd, args, argsLenAtDash))
			cmdutil.CheckErr(options.Validate())
			cmdutil.CheckErr(options.Run())
		},
		BashCompletionFunction: bashCompletionFunc,
	}

	// Set flags
	cmdutil.AddPodRunningTimeoutFlag(cmd, defaultPodExecTimeout)
	cmd.Flags().StringVarP(&options.Namespace, "namespace", "n", "default", "If present, the namespace scope for this CLI request")
	cmd.Flags().StringVarP(&options.ContainerName, "container", "c", "", "Container name. If omitted, the first container in the pod will be chosen")
	cmd.Flags().BoolVarP(&options.Stdin, "stdin", "i", false, "Pass stdin to the container")
	cmd.Flags().BoolVarP(&options.TTY, "tty", "t", false, "Stdin is a TTY")
	cmd.Flags().StringVarP(&options.Completion, "completion", "C", "", "Output shell completion code for the specified shell (bash or zsh)")

	// Set shell completion
	for name, completion := range bashCompletionFlags {
		cmd.Flag(name).Annotations = map[string][]string{}
		cmd.Flag(name).Annotations[cobra.BashCompCustom] = append(
			cmd.Flag(name).Annotations[cobra.BashCompCustom],
			completion,
		)
	}
	options.cmd = cmd

	return cmd
}

// PexecOptions
type PexecOptions struct {
	exec.ExecOptions

	restClientGetter genericclioptions.RESTClientGetter
	cmd              *cobra.Command

	Namespace  string
	Completion string
}

func (p *PexecOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, argsIn []string, argsLenAtDash int) error {
	// Print shell completion code
	if len(p.Completion) != 0 {
		if p.Completion == "bash" {
			p.cmd.GenBashCompletion(os.Stdout)
			os.Exit(0)
		} else if p.Completion == "zsh" {
			os.Stdout.Write([]byte(zshHead))
			buf := new(bytes.Buffer)
			p.cmd.GenBashCompletion(buf)
			os.Stdout.Write(buf.Bytes())
			os.Stdout.Write([]byte(zshTail))
			os.Exit(0)
		}
		return fmt.Errorf("%s is not supported shell", p.Completion)
	}

	// Check arguments
	if len(argsIn) == 0 || argsLenAtDash == 0 {
		return cmdutil.UsageErrorf(cmd, kpexecUsageStr)
	}
	p.restClientGetter = f

	return p.ExecOptions.Complete(f, cmd, argsIn, cmd.ArgsLenAtDash())
}

func (p *PexecOptions) Validate() error {
	return p.ExecOptions.Validate()
}

func (p *PexecOptions) Run() error {
	// Get pod info
	builder := p.Builder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(p.Namespace).DefaultNamespace().ResourceNames("pods", p.ResourceName)
	obj, err := builder.Do().Object()
	if err != nil {
		return err
	}
	p.Pod, err = p.ExecutablePodFn(p.restClientGetter, obj, p.GetPodTimeout)
	if err != nil {
		return err
	}

	// Check target pod status
	targetPod := p.Pod
	if targetPod.Status.Phase == corev1.PodSucceeded || targetPod.Status.Phase == corev1.PodFailed {
		return fmt.Errorf("cannot exec into a container in a completed pod; current phase is %s", targetPod.Status.Phase)
	}

	// Get target container info
	containerName := p.ContainerName
	if len(containerName) == 0 {
		if len(targetPod.Spec.Containers) > 1 {
			fmt.Fprintf(p.ErrOut, "Defaulting container name to %s.\n", targetPod.Spec.Containers[0].Name)
			if p.EnableSuggestedCmdUsage {
				fmt.Fprintf(p.ErrOut, "Use '%s describe pod/%s -n %s' to see all of the containers in this pod.\n", p.ParentCommandName, targetPod.Name, p.Namespace)
			}
		}
		containerName = targetPod.Spec.Containers[0].Name
	}
	containerRuntime, containerID, err := findContainerRuntimeID(targetPod, containerName)
	if err != nil {
		return err
	}

	// Find cnsenter daemonset pod on the node where the target container is running
	nodePods, err := p.PodClient.Pods("kube-system").
		List(metav1.ListOptions{FieldSelector: "spec.nodeName=" + targetPod.Spec.NodeName})
	if err != nil {
		return err
	}
	nodepausePod, err := findNodepausePod(nodePods.Items)
	if err != nil {
		return err
	}

	// Set terminal
	// ensure we can recover the terminal while attached
	t := p.SetupTTY()
	var sizeQueue remotecommand.TerminalSizeQueue
	if t.Raw {
		// this call spawns a goroutine to monitor/update the terminal size
		sizeQueue = t.MonitorSize(t.GetSize())

		// unset p.Err if it was previously set because both stdout and stderr go over p.Out when tty is true
		p.ErrOut = nil
	}

	// Modify command to exec cnsenter in daemonset pod
	command := []string{"cnsenter", "--uts", "--mount", "--ipc", "--net", "--pid", "--cgroup"}
	command = append(command, "--basedir="+hostDirectory)
	command = append(command, "--socket="+containerdSocketFile)
	command = append(command, "--runtime="+containerRuntime)
	command = append(command, "--container="+containerID)
	command = append(command, "--")
	command = append(command, p.Command...)

	// Exec command throw cnsenter pod
	fn := func() error {
		restClient, err := restclient.RESTClientFor(p.Config)
		if err != nil {
			return err
		}

		req := restClient.Post().
			Resource("pods").
			Name(nodepausePod.Name).
			Namespace(nodepausePod.Namespace).
			SubResource("exec")
		req.VersionedParams(&corev1.PodExecOptions{
			Container: nodepausePod.Spec.Containers[0].Name,
			Command:   command,
			Stdin:     p.Stdin,
			Stdout:    p.Out != nil,
			Stderr:    p.ErrOut != nil,
			TTY:       t.Raw,
		}, scheme.ParameterCodec)

		return p.Executor.Execute("POST", req.URL(), p.Config, p.In, p.Out, p.ErrOut, t.Raw, sizeQueue)
	}
	if err := t.Safe(fn); err != nil {
		return err
	}

	return nil
}

func findContainerRuntimeID(pod *corev1.Pod, name string) (string, string, error) {
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == name {
			u, err := url.Parse(status.ContainerID)
			if err != nil {
				return "", "", fmt.Errorf("parse container ID error")
			}
			return u.Scheme, u.Host, nil
		}
	}

	return "", "", fmt.Errorf("no container runtime, ID info")
}

func findNodepausePod(pods []corev1.Pod) (*corev1.Pod, error) {
	for _, pod := range pods {
		for _, ref := range pod.ObjectMeta.OwnerReferences {
			if ref.Kind == "DaemonSet" && ref.Name == "nodepause" {
				return &pod, nil
			}
		}
	}

	return nil, fmt.Errorf("no nodepause pod")
}
