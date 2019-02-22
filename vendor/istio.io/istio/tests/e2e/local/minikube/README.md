# Benefits:
1. Set up a local minikube VM Environment once and run E2E tests on local machine, so you can test and debug locally.
1. No need to worry about kubernetes cluster setup.

# Prereqs:
1. Set up Istio Dev envrionment using https://github.com/istio/istio/wiki/Dev-Guide.

1. Install
  * [kvm2 for linux](https://www.linux-kvm.org/page/Main_Page) 
  * [hyperkit for mac](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperkit-driver)
  * [docker](https://docs.docker.com/) - Verify `docker version` returns version >= 18.03.0-ce
  * [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) - Verify `minikube version` returns version >= minikube 0.27.0
  * [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl) - Verify `kubectl version` returns both server and client versions
  * [curl](https://curl.haxx.se/) - Verify `curl --help` prints the help information.

You can run the following script to check/install of all pre-requisites, or use it as a reference to install them manually.
(This requires installation of [Homebrew](https://brew.sh) on macOS or debian based Linux distributions)

```bash
. ./install_prereqs.sh
```

# Steps
## 1. Set up Minikube Environment
```bash
. ./setup_host.sh
```

## 2. Build istio images
Push images from your local dev environment to the local registry on host:
```bash
. ./setup_test.sh
```
You should push new images whenever you modify istio source code.

## 2. Run tests!
You can issue test commands on your host machine.
E.g.
```bash
cd $ISTIO/istio
make e2e_simple E2E_ARGS="--use_local_cluster" HUB=localhost:5000 TAG=latest
```
Note the special arguments like **E2E_ARGS**, **HUB**, and **TAG**. They are required to run these tests with the local cluster and a local registry inside the VM. And you can run multiple E2E tests sequentially against the same VM.

# Cleanup
To save the minikube status:
```bash
minikube stop
```

To destroy the minikube:
```bash
minikube delete
``` 

To cleanup host settings only (remove docker daemon setup and port forwarding)
```bash
. ./cleanup_host.sh
```

# Troubleshooting
Please refer [Troubleshooting](Troubleshooting.md) doc for information on this.
