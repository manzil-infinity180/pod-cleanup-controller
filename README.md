# pod-cleanup-controller
A controller that watches for pods in "Evicted" or "Failed" status and deletes them automatically after X minutes.

```
Use Case: A controller that watches for pods in "Evicted" or "Failed" status and deletes them automatically after X minutes.

Skills Practiced:

client-go Shared Informer (manual way) or Kubebuilder scaffolding

Use of channels, goroutines, WaitGroups for concurrent deletion

JSON logging/formatting

Leader election (optional)

---

✅ Phase 1: Pod Cleanup Controller Using Shared Informer (No Kubebuilder)
Goal: Build a lightweight controller using client-go shared informer to delete "Failed" or "Evicted" pods after X minutes.

🔧 Tools & Concepts You’ll Learn
Go routines, channels, WaitGroup

SharedInformer (no Kubebuilder)

Kubernetes client-go

JSON logging

Graceful shutdown via context and signal.Notify

🔁 Basic Flow
Start a shared informer to watch pods.

On pod ADD or UPDATE:

If status is Failed or Evicted, spawn a goroutine to wait N minutes then delete it.

Use a buffered channel to avoid memory overload.

Log actions in JSON format.
```

