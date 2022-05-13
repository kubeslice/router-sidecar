# router-sidecar


* The Slice Router is a network service component that provides a virtual L3 IP routing functionality within a cluster for the Slice overlay network.
* Each slice in a cluster has one slice router with the possibility of a redundant pair option. 
* The Slice Operator manages the lifecycle of a Slice Router by overseeing the deployment, configuration, continuous monitoring, and management of the Slice Router.
* The Slice Router provides a full mesh network connectivity between the application pods and slice gateway pods in a cluster. 

## Getting Started 
It is strongly recommended to use a released version.

For information on installing KubeSlice on kind clusters, see [getting started with kind clusters](https://docs.avesha.io/opensource/getting-started-with-kind-clusters) or try out the example script in [kind-based example](https://github.com/kubeslice/examples/tree/master/kind).

For information on installing KubeSlice on cloud clusters, see [getting started with cloud clusters](https://docs.avesha.io/opensource/getting-started-with-cloud-clusters).


### Prerequisites

* Docker installed and running in your local machine.
* A running [`kind`](https://kind.sigs.k8s.io/)
* [`kubectl`](https://kubernetes.io/docs/tasks/tools/) installed and configured
* Follow the getting started from above, to install [`kubeslice-controller`](https://github.com/kubeslice/kubeslice-controller) and [`worker-operator`](https://github.com/kubeslice/worker-operator).

# Local Build and Update 

## Latest Docker Hub Image

```console
docker pull aveshasystems/kubeslice-router-sidecar:latest
```

## Setting up Your Helm Repo

If you have not added avesha helm repo yet, add it.

```console
helm repo add avesha https://kubeslice.github.io/charts/
```

Upgrade the avesha helm repo.

```console
helm repo update
```

### Build Docker Images

1. Clone the latest version of router-sidecar from  the `master` branch.

```bash
git clone https://github.com/kubeslice/router-sidecar.git
cd router-sidecar
```

2. Adjust `VERSION` variable in the Makefile to change the docker tag to be built.
   Image is set as `docker.io/aveshasystems/router-sidecar:$(VERSION)` in the Makefile. Change this if required

```console
make docker-build
```

### Running Local Image on Kind Clusters

1. You can load the router-sidecar docker image into the kind cluster.

```bash
kind load docker-image my-custom-image:unique-tag --name clustername
```

Example

```console
kind load docker-image aveshasystems/router-sidecar:1.2.1 --name kind
```

2. Check the loaded image in the cluster. Modify the node name if required.

```console
docker exec -it <node-name> crictl images
```

Example.

```console
docker exec -it kind-control-plane crictl images
```

### Deploy in a Cluster

Update chart values file `yourvaluesfile.yaml` that you have previously created.
Refer to [values.yaml](https://github.com/kubeslice/charts/blob/master/kubeslice-worker/values.yaml) to create `yourvaluesfiel.yaml` and update the routerSidecar image subsection to use the local image.

From the sample:

```
routerSidecar:
  image: docker.io/aveshasystems/kubeslice-router-sidecar
  tag: 0.1.0
```

Change it to:

```
routerSidecar:
  image: <my-custom-image>
  tag: <unique-tag>
```

Deploy the Updated Chart

```console
make chart-deploy VALUESFILE=yourvaluesfile.yaml
```

### Verify the router-sidecar Pods are Running

```bash
kubectl get pods -n kubeslice-system | grep vl3-slice-* 
```

## License

Apache 2.0 License.
