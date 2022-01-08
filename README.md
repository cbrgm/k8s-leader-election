# K8s Leader Election Demo

This is a demo service which uses the k8s `coordination.k8s.io` API to perform leader election between running service instances (now on called "nodes") in a cluster.

Literature:

* [K8s Leader Election by Carlos Becker](https://carlosbecker.com/posts/k8s-leader-election/)

## Demo

```yaml
apiVersion: v1
automountServiceAccountToken: true
kind: ServiceAccount
metadata:
  name: leader-election-demo-sa
  labels:
    app: leader-election-demo
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: leader-election-demo-role
  labels:
    app: leader-election-demo
rules:
  - apiGroups:
      - coordination.k8s.io
    resources:
      - leases
    verbs:
      - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: leader-election-demo-rolebinding
  labels:
    app: leader-election-demo
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: leader-election-demo-role
subjects:
  - kind: ServiceAccount
    name: leader-election-demo-sa
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: leader-election-demo
  labels:
    app: leader-election-demo
spec:
  replicas: 3
  selector:
    matchLabels:
      app: leader-election-demo
  template:
    serviceAccountName: leader-election-demo-sa
    automountServiceAccountToken: true
    metadata:
      labels:
        app: leader-election-demo
    spec:
      containers:
        - name: leader-election-node
          image: quay.io/cbrgm/leader-election-node:latest
          args:
            - "--node-id=$(POD_NAME)"
            - "--leaselock.name=leader-election-demo"
            - "--leaselock.namespace=default"
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: metadata.name
```

### Leader election in action

```bash
➜  k8s-leader-election kubectl get pods

NAME                                    READY   STATUS    RESTARTS   AGE
leader-election-demo-684dc87bb8-4p4rx   1/1     Running   0          11s
leader-election-demo-684dc87bb8-7f8t7   1/1     Running   0          11s
leader-election-demo-684dc87bb8-h8hsb   1/1     Running   0          11s
```

You can also see that a new lease object has been created for our service instances

```bash
 k8s-leader-election k get leases
NAME                   HOLDER                                  AGE
leader-election-demo   leader-election-demo-684dc87bb8-h8hsb   4m36s
```

Watch the logs:


Node `leader-election-demo-684dc87bb8-h8hsb` (leader)
```bash
 k8s-leader-election k logs leader-election-demo-684dc87bb8-h8hsb 
I0108 18:56:32.463185       1 leaderelection.go:248] attempting to acquire leader lease default/leader-election-demo...
I0108 18:56:32.482107       1 leaderelection.go:258] successfully acquired lease default/leader-election-demo
I0108 18:56:32.482226       1 main.go:90] I am the leader, will do management stuff now: leader-election-demo-684dc87bb8-h8hsb
```

Node `leader-election-demo-684dc87bb8-4p4rx` (Child 1)

```bash
➜  k8s-leader-election k logs leader-election-demo-684dc87bb8-4p4rx
I0108 18:56:35.218393       1 leaderelection.go:248] attempting to acquire leader lease default/leader-election-demo...
I0108 18:56:35.233434       1 main.go:102] new leader elected: leader-election-demo-684dc87bb8-h8hsb
```

Node `leader-election-demo-684dc87bb8-7f8t7` (Child 2)

```bash
➜  k8s-leader-election k logs leader-election-demo-684dc87bb8-7f8t7 
I0108 18:56:33.854612       1 leaderelection.go:248] attempting to acquire leader lease default/leader-election-demo...
I0108 18:56:33.868191       1 main.go:102] new leader elected: leader-election-demo-684dc87bb8-h8hsb
```

When you kill the leader (`kubectl delete pod leader-election-demo-684dc87bb8-h8hsb`) and watch the logs again, you'll see that a new leader got elected ! Awesome!
