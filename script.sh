#!/bin/bash

# Define the namespaces and deployments
NAMESPACES=("default" "prod")
DEPLOYMENTS=("nginx" "nginx2")
IMAGE="nginx"

# Function to create deployments
create_deployments() {
  for ns in "${NAMESPACES[@]}"; do
    for dep in "${DEPLOYMENTS[@]}"; do
      echo "Creating deployment $dep in namespace $ns with image $IMAGE"
      kubectl create deployment "$dep" --image="$IMAGE" --namespace="$ns"
    done
  done
}

# Function to delete deployments
delete_deployments() {
  for ns in "${NAMESPACES[@]}"; do
    for dep in "${DEPLOYMENTS[@]}"; do
      echo "Deleting deployment $dep from namespace $ns"
      kubectl delete deployment "$dep" --namespace="$ns"
    done
  done
}

# Help message
help_message() {
  echo "Usage: $0 {deploy|delete|help}"
  echo "  deploy  - Create deployments in default and prod namespaces"
  echo "  delete  - Delete deployments from default and prod namespaces"
  echo "  help    - Display this help message"
}

# Main script logic
case "$1" in
  deploy)
    create_deployments
    ;;
  delete)
    delete_deployments
    ;;
  help|*)
    help_message
    ;;
esac
