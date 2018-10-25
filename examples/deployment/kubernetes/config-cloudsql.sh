DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export PROJECT_ID=trillian-cloudsql-ci
export CLUSTER_NAME=trillian
export REGION=us-east1
export MASTER_ZONE="${REGION}-b"
export NODE_LOCATIONS="${REGION}-b,${REGION}-c,${REGION}-d"
export CONFIGMAP=${DIR}/trillian-cloudsql.yaml

export POOLSIZE=2
export MACHINE_TYPE="n1-standard-2"
export STORAGE="cloudsql"
export STORAGE_MACHINE_TYPE="db-n1-standard-4"

export RUN_MAP=false
