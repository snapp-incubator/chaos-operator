# A toxiproxy chaos operator for simulate chaos experiments on Kubernetes
The Chaos Operator is designed to simulate chaos within applications and Kubernetes infrastructure, utilizing Toxiproxy under the hood. Its primary objective is to automate the installation of Toxiproxy and persist its configuration within the Kubernetes cluster.
With Chaos Operator, you can conveniently simulate various abnormalities in a controlled way that might occur in reality during development, testing, and production environments. This allows you to find potential problems in the system. Currently, it offers latency and timeout types of fault simulation.


## How the Chaos Operator Works
The Chaos Operator enables you to create proxies with various toxics that mimic real-world network issues. It creates proxy instances that sit between your application and the actual service, intercepting and controlling the network traffic. Toxics are components that introduce various network conditions or issues, simulating problems like latency, timeouts, and connection issues. Toxics are applied to specific proxies, allowing you to selectively introduce these issues for testing.


## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/chaos-operator:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/chaos-operator:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller from the cluster:

```sh
make undeploy
```


### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)


## Contributing
Contributions are welcomed and you can contribute by raising issues, improving the documentation, contributing to the core framework and tooling, etc.

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

