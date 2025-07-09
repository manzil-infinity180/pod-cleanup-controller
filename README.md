# 🧹 Pod Cleanup Controller
A controller that watches for pods in "Evicted" or "Failed" status and deletes them automatically after X minutes.

A Kubernetes controller written in Go that automatically deletes pods that are:
- In `Failed` phase
- In `Evicted` state
- In `CrashLoopBackOff` with restart count ≥ 5

This project is built using `client-go` and uses shared informers for real-time event processing.

---

## ✨ Features (IN Progress)

- Watches all pods across the cluster
- Identifies problematic pods
- Waits for a configurable delay (default 5 minutes)
- Deletes the pod after grace period
- Logs structured pod metadata in JSON
- Lightweight and stateless

---

## 📦 Use Cases (In progress)
- Automated cleanup for crashlooping test/dev pods
- Reclaiming cluster resources
- Simplifying observability during CI/CD tests
- Educational use for learning Kubernetes controllers and shared informers

---

## 🚀 Getting Started

### Prerequisites

- Go 1.20+
- Access to a Kubernetes cluster (e.g., via `kubectl`)
- [`kubectl` context](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) configured to a running cluster

---

## 🛠️ Running Locally

Clone the repo and build:

```bash
git clone https://github.com/manzil-infinity180/pod-cleanup-controller.git
cd pod-cleanup-controller

go mod tidy
go run main.go
```

---
## 🔬 How It Works
The controller uses a shared informer to watch pod changes. It checks:
1. pod.Status.Phase == Failed
2. pod.Status.Reason == "Evicted"
3. Any container with:
   - Waiting.Reason == CrashLoopBackOff
   - RestartCount >= 5

* If a pod matches, it will:
  - Sleep for 5 minutes (configurable)
  - Delete the pod
  - Log the deletion in JSON

## Example 
```yaml
# crashloop.yaml
apiVersion: v1
kind: Pod
metadata:
  name: failed-pod
spec:
  containers:
  - name: busybox
    image: busybox
    command: ["false"]
    restartPolicy: Always
```
```bash
kubectl apply -f crashloop.yaml

```
# 🙌 Acknowledgments
Built by [@manzil-infinity180](https://github.com/manzil-infinity180) for learning and exploration of Kubernetes controllers and Go internals.
