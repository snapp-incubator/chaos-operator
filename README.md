# A Toxiproxy Chaos Operator for simulating chaos experiments on Kubernetes
The Chaos Operator is designed to simulate chaos within applications and Kubernetes infrastructure, utilizing Toxiproxy under the hood. Its primary objective is to automate the installation of Toxiproxy and persist its configuration within the Kubernetes cluster.
With Chaos Operator, you can conveniently simulate various abnormalities in a controlled way that might occur in reality during development, testing, and production environments. This allows you to find potential problems in the system. Currently, it offers latency and timeout types of fault simulation.


## How the Chaos Operator Works
The Chaos Operator enables you to create proxies with various toxics that mimic real-world network issues. It creates proxy instances that sit between your application and the actual service, intercepting and controlling the network traffic. Toxics are components that introduce various network conditions or issues, simulating problems like latency, timeouts, and connection issues. Toxics are applied to specific proxies, allowing you to selectively introduce these issues for testing.

## Getting Started

* Install the chaos infrastructure components (RBAC, CRDs) and the Operator 
* Create a NetworkChaos custom resource to simulating the chaos

## Installation

#### Install Chaos Operator using Helm

[Helm](https://helm.sh) must be installed to use the charts. Please refer to
Helm's [documentation](https://helm.sh/docs) to get started.

Once Helm has been set up correctly, add the repo as follows:

```shell
helm repo add chaos-operator https://snapp-incubator.github.io/chaos-operator
```

If you had already added this repo earlier, run `helm repo update` to retrieve
the latest versions of the packages. You can then run `helm search repo
chaos-operator` to see the charts.

To install the rls-operator chart:

```shell
helm install chaos-operator chaos-operator/chaos-operator
```

To uninstall the chart:

```shell
helm delete chaos-operator
```


#### Create the namespace to install Chaos Operator

```
kubectl create ns my-chaos-operator
```

#### Verify the installation

```bash
kubectl get po -n my-chaos-operator
NAME                                                       READY   STATUS    RESTARTS          AGE
chaos-operator-controller-manager-694fcd95fc-r6kjb         2/2     Running   0                 1d
```
```bash
kubectl get sa -n chaos
NAME                                      SECRETS   AGE
chaos-operator-controller-manager         2         1d
```

## How to use
To simulate a fault for your service, you need to create a Chaos Proxy with a NetworkChaos object. Create a new YAML file to define a Chaos Proxy and add configuration parameters based on the type of Chaos you want to create.

```yaml
apiVersion: "chaos.snappcloud.io/v1alpha1"
kind: NetworkChaos
metadata:
  name: network-chaos-example
  namespace: network-chaos-example
spec:
  upstream:   
    name: my-upstream-svc
    port: "8080"
  enabled: true
  stream: upstream
  latencyToxic:
   latency: 7000
   jitter: 1
   probabilty: 1.0
  timeoutToxic:
   timeout: 2000
    probabilty: 1.0
```
This example defines a proxy targeting the my-upstream-svc service with port 8080 under the network-chaos-example namespace.
After creating the NetworkChaos object, a proxy is created and sits between your application and the my-upstream-svc service:

```bash
oc get svc 
NAME                                     TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
network-chaos-example-my-upstream-svc    ClusterIP   172.30.115.120   <none>        38861/TCP   2h
```
Now, you can request the network-chaos-example-my-upstream-svc service, and you will see the request with a faulty response.

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

