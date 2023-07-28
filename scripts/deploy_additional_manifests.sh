#!/bin/bash

# Install jq if not already installed
if ! command -v jq &>/dev/null; then
  echo "jq is not installed. Installing jq..."
  if [[ "$(uname)" == "Darwin" ]]; then
    brew install jq
  else
    sudo apt-get update && sudo apt-get install -y jq
  fi
fi

# Step 1: Get the CLUSTER variable
CLUSTER=$(az aks list --query "[0].name" -o json | jq -r '.')
echo "CLUSTER: $CLUSTER"

# Step 2: Get the RG variable
RG=$(az aks list --query "[?name=='$CLUSTER'].resourceGroup" -o json | jq -r '.[0]')
echo "RG: $RG"

# Step 3: Get the Location variable
LOCATION=$(az aks show -n $CLUSTER -g $RG --query location -o tsv)
echo "Location: $LOCATION"

# Step 4: Get AZURE_TENANT_ID
AZURE_TENANT_ID=$(az account show --query tenantId -o tsv)
echo "AZURE_TENANT_ID: $AZURE_TENANT_ID"

# Step 5: Get CLIENT_ID
CLIENT_ID=$(az aks show -g $RG -n $CLUSTER --query addonProfiles.azureKeyvaultSecretsProvider.identity.clientId -o tsv)
echo "CLIENT_ID: $CLIENT_ID"

# Step 6: Get IP address
RG2="MC_${RG}_${CLUSTER}_${LOCATION}"
IP=$(az network public-ip list --resource-group "$RG2" --query '[0].ipAddress' -o tsv)
echo "IP: $IP"

# Step 7: Apply Service manifest with IP
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: aks-demo-service
  namespace: aks-demo
spec:
  selector:
    app: aks-demo-log
  ports:
  - protocol: TCP
    port: 8081
    targetPort: 8000
  type: LoadBalancer
  loadBalancerIP: "$IP"
EOF

# Step 8: Apply SecretProviderClass manifest with CLIENT_ID
cat <<EOF | kubectl apply -f -
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: aks-demo-secret-class
  namespace: aks-demo
spec:
  provider: azure
  parameters:
    usePodIdentity: "false"
    useVMManagedIdentity: "true"
    userAssignedIdentityID: "$CLIENT_ID"
    keyvaultName: aks-demo-kv
    objects:  |
      array:
        - |
          objectName: aks-demo-kv-user
          objectType: secret
          objectVersion: ""
        - |
          objectName: aks-demo-kv-tg-token
          objectType: secret
          objectVersion: ""
        - |
          objectName: aks-demo-kv-password
          objectType: secret
          objectVersion: ""
        - |
          objectName: aks-demo-kv-database
          objectType: secret
          objectVersion: ""
        - |
          objectName: aks-demo-kv-port
          objectType: secret
          objectVersion: ""
        - |
          objectName: aks-demo-kv-server
          objectType: secret
          objectVersion: ""
    tenantId: "$AZURE_TENANT_ID"
EOF

echo "Finished applying manifests."

