## Environment Setup
### Setting up a kind cluster
Kind is a tool for running a local version of kubernetea for development use. In case you already have an operational kubernetes cluster you can skip this step. If you don't make sure you have [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) installed on your machine.
Another requirement is changing the the "storage-driver" configuration in the docker config. Docker uses OverlayFs by default,
we need to change it to vfs. In order to change this configuration you need to edit a file called "deamon.json".
Once you have it installed, run the following command in order to create a local kubernetes cluster.

```bash
kind create cluster --name demo --config ./kind-config.yaml
```

### Setting up an NFS server on the cluster
An OSRMCluster requires a volume that can be shared between multiple pods. In order to achieve that we are going to use an NFS server and a volume provisioner that will take advantage of this NFS server. Same as the previous step, you can skip this step if you already have a PersistentVolume provisioner installed on your cluster (if you are running in a cloud environment such as GCP for example).

First, we will install the NFS server:

```bash
kubectl create -f https://raw.githubusercontent.com/kubernetes-csi/csi-driver-nfs/master/deploy/example/nfs-provisioner/nfs-server.yaml
```

Next, we will install the provisioner using helm, so please make you have [helm](https://helm.sh/docs/intro/install/) installed on your machine.

```bash
helm repo add csi-driver-nfs https://raw.githubusercontent.com/kubernetes-csi/csi-driver-nfs/master/charts

helm install csi-driver-nfs csi-driver-nfs/csi-driver-nfs --namespace kube-system --version v3.1.0
```

Once you have a functional NFS server and an NFS volume provisioner you need create a new StorageClass on your cluster.

```bash
kubectl create -f https://raw.githubusercontent.com/kubernetes-csi/csi-driver-nfs/master/deploy/example/storageclass-nfs.yaml
```

### Installing an Ingress controller
This demo also create an ingress that will enable us access our OSRMCluster from outside the cluster. In this demo we will use Nginx ingress contoller (If you are running in GCP environment you can skip this step).

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

## OSRMCluster installation
### Installing the operator
The easiest way to install the operator at the moment is cloning this repository and run a *make* command (Make sure you have [golang](https://go.dev/) installed on your machine).

```bash
git clone https://github.com/itayankri/OSRM-Operator.git
make install deploy
```
This command will create a new CRD on your cluster called "osrmcluster" and a new namespace for the operator.

### Creating a new OSRMCluster
This directory contains two OSRMCluster resource examples - one for a single-profile osrm cluster and one for a multi-profile osrm cluster. You can create either on of them using kubectl.

```bash
kubectl apply -f multi_profile_osrm_cluster.yaml
```
