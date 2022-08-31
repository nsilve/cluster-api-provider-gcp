export GCP_REGION="us-central1"
export GCP_PROJECT="elastic-esp-dev"
export CONTROL_PLANE_MACHINE_COUNT=1
export WORKER_MACHINE_COUNT=1
# Make sure to use same kubernetes version here as building the GCE image
export KUBERNETES_VERSION=1.24.1
export GCP_CONTROL_PLANE_MACHINE_TYPE=n2-standard-2
export GCP_NODE_MACHINE_TYPE=n2-standard-2
export GCP_NETWORK_NAME="default"
export CLUSTER_NAME="nsilve-tilt-capg-test"