#!/bin/bash

CLUSTER_NAME="taskflow-cluster"
IMAGE_TAG="latest"
SERVICES=("auth-service" "workspace-service" "notification-service")
K8S_SERVICES=("auth" "workspace" "notification")


echo "Taskflow Orchestrator Script"


read -p "Do you want to create a Kind cluster? (y/n): " create_cluster
if [[ "$create_cluster" == "y" ]]; then
    echo "Creating cluster '$CLUSTER_NAME'..."
    kind create cluster --name "$CLUSTER_NAME"
fi

read -p "Do you want to build and load Docker images? (y/n): " build_images
if [[ "$build_images" == "y" ]]; then
    for service in "${SERVICES[@]}"; do
        echo "Building $service..."
        docker build -t "$service:$IMAGE_TAG" -f "services/$service/Dockerfile" .
        echo "Loading $service into $CLUSTER_NAME..."
        kind load docker-image "$service:$IMAGE_TAG" --name "$CLUSTER_NAME"
    done
fi

if [[ "$install_monitoring" == "y" ]]; then
    echo "Installing/Updating Monitoring stack..."
    
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
    helm repo update

    helm uninstall ingress-nginx -n ingress-nginx 2>/dev/null
    helm uninstall kube-stack -n monitoring 2>/dev/null
    helm uninstall tempo -n monitoring 2>/dev/null
    helm uninstall loki -n monitoring 2>/dev/null
    helm uninstall otel-collector -n monitoring 2>/dev/null

    helm install ingress-nginx ingress-nginx/ingress-nginx --namespace ingress-nginx --create-namespace

    helm install kube-stack prometheus-community/kube-prometheus-stack --namespace monitoring --create-namespace

    helm install tempo grafana/tempo --namespace monitoring

    helm install loki grafana/loki --namespace monitoring \
    --set deploymentMode=SingleBinary \
    --set loki.auth_enabled=false \
    --set singleBinary.replicas=1 \
    --set loki.storage.type=filesystem \
    --set loki.storage.bucketNames.chunks=chunks \
    --set loki.storage.bucketNames.ruler=ruler \
    --set loki.storage.bucketNames.admin=admin \
    --set backend.replicas=0 \
    --set read.replicas=0 \
    --set write.replicas=0 \
    --set loki.useTestSchema=true 

    helm install otel-collector open-telemetry/opentelemetry-collector --namespace monitoring \
    --set mode=deployment \
    --set image.repository="otel/opentelemetry-collector" \
    --set config.exporters.otlp.endpoint="tempo:4317" \
    --set config.exporters.otlp.tls.insecure=true

    echo "Monitoring stack fully installed."
fi

read -p "Do you want to apply Kubernetes manifests? (y/n): " apply_k8s
if [[ "$apply_k8s" == "y" ]]; then
    echo "Deploying services..."
    for service in "${K8S_SERVICES[@]}"; do
        kubectl apply -f "deployments/k8s/$service/"
    done
    echo "Deployments completed."
fi

echo "Pipeline Finished Successfully!"


# after finish run these command
# kubectl apply -f "deployments/k8s/global-ingress.yaml"
# kubectl port-forward svc/kube-stack-grafana 3000:80 -n monitoring
# kubectl port-forward svc/ingress-nginx-controller -n ingress-nginx 8080:80
# kubectl get secret --namespace monitoring kube-stack-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
# now take that password and login to grafana with localhost:3000 -> username: admin and password: you already get it
# add these into connection tab of grafana dashboard
# http://tempo.monitoring.svc.cluster.local:3200
# http://loki:3100