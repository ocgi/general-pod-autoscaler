#!/bin/bash

CA_BUNDLE=$(kubectl get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' | base64 | tr -d '\n')
echo $CA_BUNDLE


sed "s|\${CA_BUNDLE}|${CA_BUNDLE}|g" validatorconfig.yaml > tmpvalidatorconfig.yaml
