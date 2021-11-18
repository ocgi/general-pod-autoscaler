#!/bin/sh
kubectl delete apiservices v1beta1.metrics.k8s.io

kubectl get apiservices | grep metrics

kubectl apply -f kubernetes/metrics-server.yaml

kubectl get apiservices | grep metrics

