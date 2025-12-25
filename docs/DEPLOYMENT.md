# Students API - Kubernetes Deployment Guide

Deployment guide for running Students API on a local Kind (Kubernetes in Docker) cluster.

## Prerequisites

Ensure you have the following tools installed:

```bash
# Check Docker
docker --version
# Docker version 20.10.0 or higher

# Check Kind
kind --version
# kind v0.20.0 or higher

# Check kubectl
kubectl version --client
# Client Version: v1.28.0 or higher
```

### Installing Prerequisites

**Docker:**
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install docker.io
sudo usermod -aG docker $USER
```

**Kind:**
```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

**kubectl:**
```bash
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

## 🏗️ Architecture Overview

### Cluster Architecture

```
┌─────────────────────────────────────────────────────┐
│               Kind Cluster (Docker)                 │
│                                                     │
│  ┌───────────────┐         ┌──────────────┐         │
│  │ Control Plane │         │ Worker Node  │         │
│  │               │         │              │         │
│  │  - API Server │         │  - Pod 1     │         │
│  │  - Scheduler  │         │  - Pod 2     │         │
│  │  - Controller │         └──────────────┘         │
│  └───────────────┘                                  │
│         │                   ┌──────────────┐        │
│         └───────────────────│ Worker Node  │        │
│                             │              │        │
│                             │  - Storage   │        │
│                             └──────────────┘        │
│                                                     │
│  Port Mapping: 30080 → Host:30080                   │
└─────────────────────────────────────────────────────┘
```

### Application Architecture

```
External Request (localhost:30080)
        ↓
┌───────────────────────────────────┐
│      NodePort Service             │
│      (Port 30080 → 8080)          │
└───────────────────────────────────┘
        ↓
┌───────────────────────────────────┐
│   Load Balancer (Service)         │
│   Distributes across pods         │
└───────────────────────────────────┘
                 ↓
             ┌───┴───┐
             ↓       ↓
         ┌─────┐   ┌─────┐
         │Pod 1│   │Pod 2│  (2 Replicas for HA)
         └─────┘   └─────┘
             │       │
             └───┬───┘
                 ↓
┌───────────────────────────────────┐
│  Persistent Volume (SQLite DB)    │
└───────────────────────────────────┘
```

## 📁 Deployment Files

```
students_api/
├── Dockerfile                 # Multi-stage production build
├── .dockerignore             # Docker build optimization
├── k8s/                      # Kubernetes manifests
│   ├── kind-config.yaml      # Cluster configuration
│   ├── namespace.yaml        # Resource isolation
│   ├── configmap.yaml        # App configuration
│   ├── pvc.yaml             # Persistent storage
│   ├── deployment.yaml       # App deployment
│   └── service.yaml          # Service exposure
└── DEPLOYMENT.md             # This file
```

## Deployment
### Step 1: Create Kind Cluster

```bash
kind create cluster --config k8s/kind-config.yaml
```

**What this creates:**
- 1 Control plane node
- 2 Worker nodes
- Port mapping: 30080 (NodePort)

### Step 2: Build Docker Image

```bash
docker build -t students-api:latest .
```

**Build features:**
- Multi-stage build (builder + runtime)
- Alpine Linux base (minimal size)
- Non-root user (security)
- Health checks included

### Step 3: Load Image into Kind

```bash
kind load docker-image students-api:latest --name students-api-cluster
```

### Step 4: Apply Kubernetes Manifests

```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Apply configurations
kubectl apply -f k8s/configmap.yaml

# Create storage
kubectl apply -f k8s/pvc.yaml

# Deploy application
kubectl apply -f k8s/deployment.yaml

# Expose service
kubectl apply -f k8s/service.yaml
```

### Step 5: Verify Deployment

```bash
# Check all resources
kubectl get all -n students-api

# Wait for deployment
kubectl rollout status deployment/students-api -n students-api

# Check logs
kubectl logs -f -l app=students-api -n students-api
```

## Features

### 1. High Availability
- **2 Replicas**: Multiple pods for redundancy
- **Rolling Updates**: Zero-downtime deployments
- **Anti-Affinity**: Pods distributed across nodes

### 2. Resource Management
```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "500m"
```

### 3. Health Checks

**Liveness Probe**: Restarts unhealthy containers
```yaml
livenessProbe:
  httpGet:
    path: /
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30
```

**Readiness Probe**: Routes traffic only to ready pods
```yaml
readinessProbe:
  httpGet:
    path: /
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

**Startup Probe**: Handles slow-starting containers
```yaml
startupProbe:
  httpGet:
    path: /
    port: 8080
  failureThreshold: 10
  periodSeconds: 5
```

### 4. Security

- **Non-root user**: Runs as UID 1000
- **Read-only filesystem**: Where possible
- **Security context**: Drops all capabilities
- **ConfigMap**: Separate config from code
- **No privilege escalation**

### 5. Persistent Storage

- **PersistentVolumeClaim**: 1Gi storage
- **ReadWriteOnce**: Single pod write access
- **SQLite database**: Persists across restarts

### 6. Graceful Shutdown

- **Termination grace period**: 40 seconds
- **Shutdown timeout**: 30 seconds (from config)
- **Proper signal handling**: SIGTERM/SIGINT

## 🧪 Testing

### Basic Health Check

```bash
curl http://localhost:30080/
```

### API Endpoints

```bash
# List all students
curl http://localhost:30080/students

# Get specific student
curl http://localhost:30080/students/1

# Create student
curl -X POST http://localhost:30080/students \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Johnson",
    "email": "alice@example.com",
    "age": 22
  }'

# Test slow endpoint (for timeout testing)
curl http://localhost:30080/slow
```

## Debugging

### View Logs

```bash
# All pods
kubectl logs -f -l app=students-api -n students-api

# Specific pod
kubectl logs -f <pod-name> -n students-api

# Previous container (if crashed)
kubectl logs <pod-name> -n students-api --previous

# Last 50 lines
kubectl logs --tail=50 -l app=students-api -n students-api
```

### Pod Information

```bash
# List pods with details
kubectl get pods -n students-api -o wide

# Describe pod
kubectl describe pod <pod-name> -n students-api

# Get pod YAML
kubectl get pod <pod-name> -n students-api -o yaml
```

### Execute Commands in Pod

```bash
# Get shell access
POD=$(kubectl get pod -l app=students-api -n students-api -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it $POD -n students-api -- sh

# Check database
kubectl exec -it $POD -n students-api -- ls -la /var/lib/students_api/

# Check processes
kubectl exec -it $POD -n students-api -- ps aux
```

### Events

```bash
# Get events in namespace
kubectl get events -n students-api --sort-by='.lastTimestamp'

# Watch events in real-time
kubectl get events -n students-api --watch
```

### Resource Usage

```bash
# Pod resource usage (requires metrics-server)
kubectl top pods -n students-api

# Node resource usage
kubectl top nodes
```

## 📈 Scaling

### Manual Scaling

```bash
# Scale to 3 replicas
kubectl scale deployment students-api --replicas=3 -n students-api

# Scale to 5 replicas
kubectl scale deployment students-api --replicas=5 -n students-api

# Scale down to 1
kubectl scale deployment students-api --replicas=1 -n students-api

# Watch scaling
kubectl get pods -n students-api --watch
```

### Auto-Scaling (HPA)

Create HorizontalPodAutoscaler:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: students-api-hpa
  namespace: students-api
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: students-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

Apply:
```bash
kubectl apply -f k8s/hpa.yaml
kubectl get hpa -n students-api
```

## 🧹 Cleanup

### Manual Cleanup Process

#### Option 1: Delete Namespace Only (Keep Cluster)

This removes all application resources but keeps the Kind cluster running:

```bash
# Delete the namespace (this removes all resources in it)
kubectl delete namespace students-api

# Verify deletion
kubectl get namespaces
```

**What gets deleted:**
- Deployment (and all pods)
- Service
- ConfigMap
- PersistentVolumeClaim
- All other resources in the namespace

#### Option 2: Delete Entire Kind Cluster

This removes everything including the cluster:

```bash
# Delete the Kind cluster
kind delete cluster --name students-api-cluster

# Verify cluster is deleted
kind get clusters
```

**What gets deleted:**
- The entire Kind cluster
- All nodes (control plane + workers)
- All namespaces and resources
- Persistent volumes

#### Option 3: Delete Docker Image

Remove the Docker image from your local machine:

```bash
# Delete the image
docker rmi students-api:latest

# Or force delete if needed
docker rmi -f students-api:latest

# Verify image is deleted
docker images | grep students-api
```

#### Option 4: Complete Cleanup (Everything)

Remove everything in the correct order:

```bash
# Step 1: Delete the namespace
kubectl delete namespace students-api --timeout=60s

# Step 2: Delete the Kind cluster
kind delete cluster --name students-api-cluster

# Step 3: Delete the Docker image
docker rmi students-api:latest

# Step 4: Clean up any dangling images (optional)
docker image prune -f

# Step 5: Verify everything is cleaned up
echo "Checking remaining resources..."
kind get clusters
docker images | grep students-api
echo "Cleanup complete!"
```

### Troubleshooting Cleanup Issues

#### Namespace Stuck in "Terminating" State

If namespace deletion hangs:

```bash
# Check what's blocking deletion
kubectl get all -n students-api
kubectl get pvc -n students-api
kubectl get events -n students-api

# Force delete namespace (if needed)
kubectl delete namespace students-api --force --grace-period=0

# If still stuck, remove finalizers (advanced)
kubectl get namespace students-api -o json > namespace.json
# Edit namespace.json and remove "finalizers" section
kubectl replace --raw "/api/v1/namespaces/students-api/finalize" -f namespace.json
```

#### PersistentVolumeClaim Won't Delete

```bash
# Check PVC status
kubectl get pvc -n students-api

# Check what's using the PVC
kubectl describe pvc students-api-storage -n students-api

# Delete pods first (if PVC is in use)
kubectl delete deployment students-api -n students-api --force --grace-period=0

# Then delete PVC
kubectl delete pvc students-api-storage -n students-api --force --grace-period=0
```

#### Kind Cluster Won't Delete

```bash
# List all Kind clusters
kind get clusters

# Force delete using Docker
docker ps -a | grep students-api-cluster

# Remove containers forcefully
docker ps -a | grep students-api-cluster | awk '{print $1}' | xargs docker rm -f

# Clean up networks
docker network ls | grep kind
docker network prune -f
```

#### Docker Image in Use

```bash
# Check if image is being used by containers
docker ps -a | grep students-api

# Stop and remove containers using the image
docker ps -a | grep students-api | awk '{print $1}' | xargs docker stop
docker ps -a | grep students-api | awk '{print $1}' | xargs docker rm

# Now delete the image
docker rmi students-api:latest

# Force delete if still fails
docker rmi -f students-api:latest
```

### Quick Cleanup Commands

```bash
# Quick namespace cleanup
kubectl delete namespace students-api

# Quick cluster cleanup
kind delete cluster --name students-api-cluster

# Quick image cleanup
docker rmi students-api:latest

# Quick complete cleanup (one-liner)
kubectl delete namespace students-api; kind delete cluster --name students-api-cluster; docker rmi students-api:latest
```

### Verify Cleanup

After cleanup, verify everything is removed:

```bash
# Check namespaces
kubectl get namespaces | grep students-api
# Should return nothing

# Check Kind clusters
kind get clusters
# Should not show students-api-cluster

# Check Docker images
docker images | grep students-api
# Should return nothing

# Check Docker containers
docker ps -a | grep students-api
# Should return nothing

# Check Docker networks (optional)
docker network ls | grep kind
# Should not show students-api-cluster network
```

### Cleanup Best Practices

1. **Always delete namespace first** - This ensures proper resource cleanup
2. **Wait for graceful termination** - Give pods time to shut down properly (30-60s)
3. **Check for stuck resources** - Use `kubectl get all -n students-api` to verify
4. **Delete cluster after namespace** - Ensures no orphaned resources
5. **Clean Docker images last** - In case you need to redeploy quickly

### Partial Cleanup (For Development)

If you want to keep the cluster but redeploy:

```bash
# Delete only the deployment (keeps namespace, PVC, etc.)
kubectl delete deployment students-api -n students-api

# Or restart deployment without deletion
kubectl rollout restart deployment/students-api -n students-api

# Delete and recreate specific resources
kubectl delete -f k8s/deployment.yaml
kubectl apply -f k8s/deployment.yaml
```