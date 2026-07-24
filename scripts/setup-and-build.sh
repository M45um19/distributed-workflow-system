#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

CLUSTER_NAME="taskflow-cluster"
IMAGE_TAG="latest"
SERVICES=("auth-service" "workspace-service" "notification-service")
K8S_SERVICES=("auth" "workspace" "notification")

echo "=========================================================="
echo "          TaskFlow Local Orchestrator Setup Script         "
echo "=========================================================="
echo ""

# 1. Prerequisite Checks
check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo "Error: '$1' is required but not installed."
        echo "Please install it and try again."
        exit 1
    fi
}

echo "Checking system prerequisites..."
check_command docker
check_command kind
check_command kubectl
check_command helm
echo "All prerequisites (docker, kind, kubectl, helm) are present!"
echo ""

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    echo "Error: Docker daemon is not running. Please start Docker Desktop and run this script again."
    exit 1
fi

# 2. Database Infrastructure Setup (Docker Compose)
read -p "Do you want to spin up local database engines (Postgres, Mongo, Redis, Kafka, Temporal) via Docker Compose? (y/n): " start_compose
if [[ "$start_compose" == "y" ]]; then
    echo "Starting background databases..."
    docker compose -f deployments/docker-compose.yaml up -d
    echo "Databases started in background!"
    echo ""
fi

# 3. Kubernetes Cluster Setup (KinD)
read -p "Do you want to create/recreate the KinD Kubernetes cluster? (y/n): " create_cluster
if [[ "$create_cluster" == "y" ]]; then
    # Delete existing cluster if it exists
    if kind get clusters | grep -q "$CLUSTER_NAME"; then
        echo "Deleting existing cluster '$CLUSTER_NAME'..."
        kind delete cluster --name "$CLUSTER_NAME"
    fi
    
    echo "Creating cluster '$CLUSTER_NAME'..."
    kind create cluster --name "$CLUSTER_NAME"
    echo "KinD cluster created!"
    echo ""
fi

# 4. Build and Load Docker Images
read -p "Do you want to build microservice Docker images and load them into KinD? (y/n): " build_images
if [[ "$build_images" == "y" ]]; then
    for service in "${SERVICES[@]}"; do
        echo "----------------------------------------"
        echo "Building image for: $service"
        echo "----------------------------------------"
        docker build -t "$service:$IMAGE_TAG" -f "services/$service/Dockerfile" .
        
        echo "Loading $service:$IMAGE_TAG into KinD cluster..."
        kind load docker-image "$service:$IMAGE_TAG" --name "$CLUSTER_NAME"
    done
    echo "All images built and loaded successfully!"
    echo ""
fi

# 5. Deploy Monitoring & Ingress Stack
read -p "Do you want to install/reinstall the Monitoring & Ingress stack (Helm)? (y/n): " install_monitoring
if [[ "$install_monitoring" == "y" ]]; then
    echo "Configuring Helm repositories..."
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
    helm repo update

    echo "Cleaning up any existing Helm deployments..."
    helm uninstall ingress-nginx -n ingress-nginx 2>/dev/null || true
    helm uninstall kube-stack -n monitoring 2>/dev/null || true
    helm uninstall tempo -n monitoring 2>/dev/null || true
    helm uninstall loki -n monitoring 2>/dev/null || true
    helm uninstall promtail -n monitoring 2>/dev/null || true
    helm uninstall otel-collector -n monitoring 2>/dev/null || true

    echo "Installing Nginx Ingress Controller..."
    helm install ingress-nginx ingress-nginx/ingress-nginx --namespace ingress-nginx --create-namespace

    echo "Installing Prometheus Operator Stack..."
    helm install kube-stack prometheus-community/kube-prometheus-stack --namespace monitoring --create-namespace

    echo "Installing Tempo (Distributed Tracing)..."
    helm install tempo grafana/tempo --namespace monitoring

    echo "Installing Loki (Log Aggregation)..."
    helm install loki grafana/loki --namespace monitoring \
        --set deploymentMode=SingleBinary \
        --set loki.auth_enabled=false \
        --set singleBinary.replicas=1 \
        --set singleBinary.resources.requests.memory=100Mi \
        --set singleBinary.resources.limits.memory=500Mi \
        --set chunksCache.enabled=false \
        --set resultsCache.enabled=false \
        --set lokiCanary.enabled=false \
        --set loki.storage.type=filesystem \
        --set loki.storage.bucketNames.chunks=chunks \
        --set loki.storage.bucketNames.ruler=ruler \
        --set loki.storage.bucketNames.admin=admin \
        --set backend.replicas=0 \
        --set read.replicas=0 \
        --set write.replicas=0 \
        --set loki.useTestSchema=true 

    echo "Installing Promtail (Log Shipping Agent)..."
    helm install promtail grafana/promtail --namespace monitoring \
        --set "config.clients[0].url=http://loki-gateway/loki/api/v1/push"

    echo "Installing OpenTelemetry Collector..."
    helm install otel-collector open-telemetry/opentelemetry-collector --namespace monitoring \
        -f deployments/k8s/otel-collector-values.yaml

    echo "Monitoring & Ingress stack installation completed."
    echo ""
fi

# 6. Deploy Microservices Manifests
read -p "Do you want to apply Kubernetes service manifests and Ingress configurations? (y/n): " apply_k8s
if [[ "$apply_k8s" == "y" ]]; then
    echo "Applying microservice manifests..."
    for service in "${K8S_SERVICES[@]}"; do
        echo "Deploying manifests in deployments/k8s/$service/..."
        kubectl apply -f "deployments/k8s/$service/"
    done
    
    echo "Applying global Ingress rules..."
    kubectl apply -f "deployments/k8s/global-ingress.yaml"
    
    echo "All manifests and ingress rules applied!"
    echo ""
fi

echo "=========================================================="
echo "          Orchestration Completed Successfully!           "
echo "=========================================================="
echo ""
echo "Follow these instructions to run and inspect your project:"
echo ""

# Fetch Grafana Admin Password
if kubectl get secret --namespace monitoring kube-stack-grafana &>/dev/null; then
    GRAFANA_PASSWORD=$(kubectl get secret --namespace monitoring kube-stack-grafana -o jsonpath="{.data.admin-password}" | base64 --decode || echo "")
fi

echo "----------------------------------------------------------"
echo "1. PORT FORWARDING (Run these commands in separate terminals)"
echo "----------------------------------------------------------"
echo "  # Port-forward the Ingress Controller (Allows Postman/Client access on port 8080)"
echo "  kubectl port-forward svc/ingress-nginx-controller -n ingress-nginx 8080:80"
echo ""
echo "  # Port-forward Grafana (For Traces, Metrics, and Logs on port 3000)"
echo "  kubectl port-forward svc/kube-stack-grafana 3000:80 -n monitoring"
echo ""

echo "----------------------------------------------------------"
echo "2. GRAFANA DASHBOARD & TRACING ACCESS"
echo "----------------------------------------------------------"
echo "  URL:       http://localhost:3000"
echo "  Username:  admin"
if [[ -n "$GRAFANA_PASSWORD" ]]; then
echo "  Password:  $GRAFANA_PASSWORD"
else
echo "  Password:  (Fetch with: kubectl get secret -n monitoring kube-stack-grafana -o jsonpath=\"{.data.admin-password}\" | base64 --decode)"
fi
echo ""
echo "  Data Sources Setup (Connections -> Data sources):"
echo "  - Add Tempo: http://tempo.monitoring.svc.cluster.local:3200"
echo "  - Add Loki:  http://loki-gateway.monitoring.svc.cluster.local"
echo ""

echo "----------------------------------------------------------"
echo "3. API & POSTMAN VERIFICATION (Hit via Port-Forwarded Ingress)"
echo "----------------------------------------------------------"
echo "  Send requests to the following endpoints on port 8080:"
echo "  - Auth Service Health:        GET  http://localhost:8080/api/v1/auth/health"
echo "  - Workspace Service Health:   GET  http://localhost:8080/api/v1/workspace/health"
echo "  - Notification Service Health:GET  http://localhost:8080/api/v1/notification/health"
echo "----------------------------------------------------------"
echo ""