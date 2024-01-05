#!/bin/bash

# Delete all SecurityIntent resources
kubectl delete securityintents --all --all-namespaces

# Delete all SecurityIntentBinding resources
kubectl delete securityintentbindings --all --all-namespaces

# Delete all NimbusPolicy resources
kubectl delete nimbuspolicies --all --all-namespaces

echo "All resources have been successfully deleted."
