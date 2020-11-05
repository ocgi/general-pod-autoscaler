#!/bin/bash

tmpdir=$(mktemp -d)

bash create-cert.sh --tmpdir ${tmpdir} --service gpa-validator --namespace kube-system --secret gpa-secret

bash patch-bundle.sh --tmpdir ${tmpdir}

kubectl apply -f ./kubernetes