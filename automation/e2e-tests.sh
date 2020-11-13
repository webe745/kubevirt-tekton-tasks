#!/usr/bin/env bash

set -ex

export SCOPE="${SCOPE:-cluster}"
export STORAGE_CLASS="${STORAGE_CLASS:-}"
export DEPLOY_NAMESPACE="${DEPLOY_NAMESPACE:-e2e-tests-$(shuf -i10000-99999 -n1)}"
export NUM_NODES=${NUM_NODES:-2}

./automation/e2e-deploy-resources.sh

oc get namespaces -o name | grep -Eq "^namespace/$DEPLOY_NAMESPACE$" || oc new-project "$DEPLOY_NAMESPACE"

export TARGET_NAMESPACE="$DEPLOY_NAMESPACE"
oc project $DEPLOY_NAMESPACE

if [[ "$SCOPE" == "cluster" ]]; then
  make deploy-dev
else
  make deploy-dev-namespace
fi


# Wait for kubevirt to be available
oc wait -n kubevirt kv kubevirt --for condition=Available --timeout 15m

make cluster-test
