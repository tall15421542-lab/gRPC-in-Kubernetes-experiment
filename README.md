# GRPC load balancing in k8s cluster: Experiment
## Pre-reading
* [GRPC load balancing in k8s cluster](https://hackmd.io/6Q-UKmWcQuqx6I5NkIUQlQ?view)
## API
### /experiment/client?type={client_type}&n={concurrent_reqs}
* client_type: the grpc client we use to send data to greeter server
    * Client: Create new channel for every request
    * ClientWithReuseConn: Create channel initially and reuse that channel for further requests.
    * ClientWithKubeResolver: Use [kubeResolver](https://github.com/sercand/kuberesolver/tree/master) as a grpc name resolver. It would monitor the Greeter-service and resolve the ips behind it dynamically.
* concurrent_reqs: It's the number of concurrent requests we send to greeter-server in the experiment.
## Architecture
![](https://hackmd.io/_uploads/r1uCU3Xqn.png)
### KubeResolver
![](https://hackmd.io/_uploads/ryQ4inQ5h.png)

## Prerequiste
* [minikube](https://minikube.sigs.k8s.io/docs/start/): A local kubernetes control plane.
* For kubeResolver, we should give service account in the greeter-client pod `GET` and `WATCH` access for resource `endpoints`.

## How to start the greeter-client & greeter-server
1. Start cluster: `minikube start`
2. (Optional) Enable local build registry
    * Run `eval $(minikube -p minikube docker-env)`, point your terminal’s docker-cli to the Docker Engine inside minikube. ([minikube handbook](https://minikube.sigs.k8s.io/docs/handbook/pushing/#1-pushing-directly-to-the-in-cluster-docker-daemon-docker-env))
3. Build greeter-client & greeter-server image: `make build-image`
4. Deploy server: `kubectl apply -f k8s/server.yaml`
5. Deploy client: `kubectl apply -f k8s/client.yaml`
    * For kubeResolver, we should also give `default` service account RBAC access for `endpoints`: `kubectl apply -f k8s/client-rbac.yaml`
6. Open tunnel to the greeter-client NodePort: in the other terminal, run `minikube service greeter-client`
7. Run experiment: `curl "localhost:{port_from_step_7}/?type={client_type}&n={concurrent_reqs}"`

## Experiment
### Client v.s ClientWithReuseConn


|  | Client | ClientWithReuseConn |
| -------- | -------- | -------- |
| n = 1000     | 687.692001ms    | 522.581208ms    |
| n = 2000     | 1.036497709s     | 547.844917ms     |
| n = 3000     | Crash     | 558.110917ms     |
| n = 10000     | Crash     | 735.631459ms     |

#### LoadBalancing
* Client
```
"count_per_ip":{
  "10.244.0.103":684,
  "10.244.0.105":662,
  "10.244.0.113":654
}
```
* ClientWithReuseConn
```
"count_per_ip":{"10.244.0.113":2000}
```

#### Comment
1. Reconnect for every request is slow. If we ignore the server sleep time 500ms, `Client` takes 8.5x~10x time than `ClientWithReuseConn`.
2. For 3000 concurrent requests, `Client` would crash due to the resource surge, while it only takes little overhead for `ClientWithReuseConn`
3. `Client` load balancing policy is more complicated than round-robin, but it's still closed to evenly distributed in our workload
4. `ClientWithReuseConn` does not have load-balancing at all.

### ClientWithResueConn v.s ClientWithKubeResolver(Replica=3)
|  | ClientWithReuseConn | ClientWithKubeResolver |
| -------- | -------- | -------- |
| n = 1000     |522.581208ms    | 513.53575ms |
| n = 2000     | 547.844917ms   |  548.33675ms|
| n = 3000     | 558.110917ms   |  533.248667ms|
| n = 10000     | 735.631459ms  |  699.490834ms |
#### LoadBalancing
* ClientWithReuseConn
```
"count_per_ip":{"10.244.0.113":10000}
```
* ClientWithKubeResolver
```
"count_per_ip":{
  "10.244.0.103":3334,
  "10.244.0.105":3333,
  "10.244.0.113":3333
}
```
#### Comment
1. HTTP/2 and Protobuf is efficient: Efficiency of one connection is as good as load-balancing in our workload. It does not need to open an new connection even when we have 10000 concurrent request.
2. KubeResolver successfully load balancing using round-robin, but it would not consider the server state.

### ClientWithKubeResolver with scaling up or down
#### Scale up to 10 replica
* `kubectl scale deployments/greeter-server --replicas=10`
* During the process of scaling
```
"count_per_ip":{
  "10.244.0.103":500,
  "10.244.0.105":500,
  "10.244.0.113":500,
  "10.244.0.122":500
}
```
* After Scaling is done
```
"count_per_ip":{
  "10.244.0.103":200,
  "10.244.0.105":200,
  "10.244.0.113":200,
  "10.244.0.121":200,
  "10.244.0.122":200,
  "10.244.0.123":200,
  "10.244.0.124":200,
  "10.244.0.125":200,
  "10.244.0.126":200,
  "10.244.0.127":200
}
```
#### Kill one pod
* Before we kill a pod
```
"count_per_ip":{
  "10.244.0.103":666,
  "10.244.0.105":667,
  "10.244.0.113":667
}
```
* After we kill a pod
    * `kubectl delete pod {pod_name}`
    * We can see one of ips changes from "10.244.0.105" to "10.244.0.128"
```
"count_per_ip":{
  "10.244.0.103":667,
  "10.244.0.113":666,
  "10.244.0.128":667
}
```

#### Scale down to 2 replica
* `kubectl scale deployments/greeter-server --replicas=2`
```
"count_per_ip":{
  "10.244.0.103":1000,
  "10.244.0.113":1000
}
```
