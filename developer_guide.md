# Development Guidelines for Router Sidecar

* The Slice Router is a network service component that provides a virtual L3 IP routing functionality within a cluster for the slice overlay network.
* Each slice in a cluster has one slice router with the possibility of a redundant pair option.
* The Slice Operator manages the lifecycle of a Slice Router by overseeing the deployment, configuration, continuous monitoring, and management of the Slice Router.
* The Slice Router provides a full mesh network connectivity between the application pods and slice gateway pods in a cluster.

## Building and Installing `router-sidecar` in a Local Kind Cluster
For more information, see [getting started with kind clusters](https://docs.avesha.io/documentation/open-source/0.2.0/getting-started-with-kind-clusters).

### Setting up Development Environment

* Go (version 1.17 or later) installed and configured in your machine ([Installing Go](https://go.dev/dl/))
* Docker installed and running in your local machine
* A running [`kind`](https://kind.sigs.k8s.io/)  cluster
* [`kubectl`](https://kubernetes.io/docs/tasks/tools/) installed and configured
* Follow the getting started from above, to install [kubeslice-controller](https://github.com/kubeslice/kubeslice-controller) 



### Building Docker Images

1. Clone the latest version of kubeslice-controller from  the `master` branch.

```bash
git clone https://github.com/kubeslice/router-sidecar.git
cd router-sidecar
```

2. Adjust image name variable `IMG` in the [`Makefile`](Makefile) to change the docker tag to be built.
   The default image is set as `IMG ?= aveshasystems/router-sidecar:${VERSION}`. Modify this if required.

```bash
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
Refer to [values.yaml](https://github.com/kubeslice/charts/blob/master/charts/kubeslice-worker/values.yaml) to create `yourvaluesfiel.yaml` and update the routerSidecar image subsection to use the local image.

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
Example output

```
vl3-slice-router-red-5b9df8d4dd-hkgkj      2/2     Running   0          26m
```

### Uninstalling the kubeslice-worker

Refer to the [uninstall guide](https://docs.avesha.io/documentation/open-source/0.2.0/getting-started-with-kind-clusters)

1. [Offboard](https://docs.avesha.io/documentation/open-source/0.2.0/getting-started-with-cloud-clusters/uninstalling-kubeslice/offboarding-namespaces) the namespaces from the slice.

2. [Delete](https://docs.avesha.io/documentation/open-source/0.2.0/getting-started-with-cloud-clusters/uninstalling-kubeslice/deleting-the-slice) the slice.

3. On the worker cluster, undeploy the kubeslice-worker charts.

```bash
# uninstall all the resources
make chart-undeploy
```

## License

Apache License 2.0
