# kpexec
kpexec is a kubernetes cli that runs commands in a container with privilege.

## Install

### Install kpexec binary
~~~
$ go install github.com/ssup2/kpexec/cmd/kpexec
~~~

### Deploy nodepause daemonset
~~~
$ kubectl apply -f https://raw.githubusercontent.com/ssup2/kpexec/master/deployments/nodepause.yaml
~~~

### Set auto completion (Optional)

#### Bash
~~~
$ source <(kpexec -C bash) 
$ echo 'source <(kpexec -C bash)' >>~/.bashrc
~~~

#### Zsh
~~~
$ source <(kpexec -C zsh) 
$ echo 'source <(kpexec -C zsh)' >>~/.zshrc
~~~

## Usage
~~~
$ kpexec [-n namespace] (POD | TYPE/NAME) [-c CONTAINER] [flags] -- COMMAND [args...]
~~~

## How it works
![kpexec Architecture](image/kpexec_Architecture.PNG)
