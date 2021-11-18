#!/bin/bash

tmpdir=$(mktemp -d)

echo "install metrics server"
bash install-metrics-server.sh

echo "create cert"
bash create-cert.sh --tmpdir ${tmpdir} --service gpa-validator --namespace kube-system --secret gpa-secret

echo "patch bundle"
bash patch-bundle.sh --tmpdir ${tmpdir}

echo "apply yaml"
kubectl apply -f ./kubernetes
