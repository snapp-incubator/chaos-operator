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

#### Create the namespace to install Chaos Operator

```
kubectl create ns chaos
```

#### Verify the installation

```bash
kubectl get po -n chaos
NAME                                                       READY   STATUS    RESTARTS          AGE
chaos-operator-controller-manager-694fcd95fc-r6kjb         2/2     Running   0                 1d
```
```bash
kubectl get sa -n chaos
NAME                                      SECRETS   AGE
chaos-operator-controller-manager         2         1d
```



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

