# kpexec
**kpexec** is a kubernetes cli that runs commands in a container with high privilege. kubectl-exec runs the command with the same privileges as the container. For example, if a container does not have network privileges, the command executed by kubectl-exec also has no network privileges. Also, kubectl-exec does not provide an option to force command to run with high privileges. This makes debugging the pod difficult. kpexec helps execute pod debugging smoothly by executing commands with high privileges regardless of container privileges.

## Install

### Install kpexec binary
Install kpexec throw `go install` through the following command.
~~~
$ go install github.com/ssup2/kpexec/cmd/kpexec
~~~

### Deploy nodepause daemonset
kpexec is completely dependent on the nodepause daemonset. Therefore, nodepause daemonset must be set through the following command.
~~~
$ kubectl apply -f https://raw.githubusercontent.com/ssup2/kpexec/master/deployments/nodepause.yaml
~~~

### Set shell autocompletion (Optional)
kpexec support shell autocompletion on bash or zsh shell base on kubectl shell autocompletion. Before setting kpexec shell autocompletion, set kubectl shell autocompletion via the link below.
* https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion

#### Bash
Set kpexec shell autocompletion to bash shell the following commands.
~~~
$ source <(kpexec -C bash) 
$ echo 'source <(kpexec -C bash)' >>~/.bashrc
~~~

#### Zsh
Set kpexec shell autocompletion to zsh shell the following commands.
~~~
$ source <(kpexec -C zsh) 
$ echo 'source <(kpexec -C zsh)' >>~/.zshrc
~~~

## Usage
kpexec provides similar usage as kubectl-exec. The following shows how to use kpexec.
~~~
$ kpexec [-n NAMESPACE] [-i] [-t] (POD | TYPE/NAME) [-c CONTAINER] -- COMMAND [args...]
~~~

## How it works
![kpexec Architecture](image/kpexec_Architecture.PNG)
The figure above shows the architecture of kpexec. Nodepause daemonset container use host (node) PID namespace to run command to container namespace. And it uses host root volume to communicate containerd and inspect host filesystem. Nodepause daemonset container use pause command as an init process such as pause container to reduce overhead. kpexec receives stdout/stderr of command and sends user's input to stdin of command throw cnsenter command in nodepause pod. cnsenter gets container info from containerd and run command in a container namespace throw nsetner command.
