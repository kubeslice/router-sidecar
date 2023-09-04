# router-sidecar

* The Slice Router is a network service component that provides a virtual L3 IP routing functionality within a cluster for the Slice overlay network.
* Each slice in a cluster has one slice router with the possibility of a redundant pair option. 
* The Slice Operator manages the lifecycle of a Slice Router by overseeing the deployment, configuration, continuous monitoring, and management of the Slice Router.
* The Slice Router provides a full mesh network connectivity between the application pods and slice gateway pods in a cluster. 

## Get Started 

Please refer to our documentation on:
- [Install KubeSlice on cloud clusters](https://kubeslice.io/documentation/open-source/1.1.0/category/install-kubeslice).

Try our the example script in [kind-based example](https://github.com/kubeslice/examples/tree/master/kind).

### Prerequisites
Before you begin, make sure the following prerequisites are met:
* Docker is installed and running on your local machine.
* A running [`kind`](https://kind.sigs.k8s.io/).
* [`kubectl`](https://kubernetes.io/docs/tasks/tools/) is installed and configured.
* You have prepared the environment to install [`kubeslice-controller`](https://github.com/kubeslice/kubeslice-controller) on the controller cluster
 and [`worker-operator`](https://github.com/kubeslice/worker-operator) on the worker cluster. For more information, see [Prerequisites](https://kubeslice.io/documentation/open-source/1.1.0/category/prerequisites).

# Build and Deploy Router Sidecar on a Kind Cluster 

To download the latest router-sidecar docker hub image, click [here](https://hub.docker.com/r/aveshasystems/kubeslice-router-sidecar).

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

2. Edit the `VERSION` variable in the Makefile to change the docker tag to be built.
   The image is set as `docker.io/aveshasystems/router-sidecar:$(VERSION)` in the Makefile. Modify this if required.

   ```console
   make docker-build
   ```

### Run Local Image on a Kind Cluster

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

### Deploy Router Sidecar on a Cluster

Update the chart values file, `yourvaluesfile.yaml` that you have previously created.
Refer to the [values.yaml](https://github.com/kubeslice/charts/blob/master/charts/kubeslice-worker/values.yaml) to create `yourvaluesfiel.yaml` and update the routerSidecar image subsection to use the local image.

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

### Verify the Installation
Verify the installation by checking the status of router-sidecar pods belonging to the `kubeslice-system` namespace.

```bash
kubectl get pods -n kubeslice-system | grep vl3-slice-* 
```
Example output

```
vl3-slice-router-red-5b9df8d4dd-hkgkj      2/2     Running   0          26m
```

## License

Apache 2.0 License.
