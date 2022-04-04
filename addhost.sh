#! /bin/bash


kubectl delete cm a-mycat-cm
kubectl apply -f og-a.yaml
kubectl apply -f og-b.yaml
