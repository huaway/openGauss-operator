#! /bin/bash

kubectl delete -f og-b.yaml
kubectl delete cm a-mycat-cm
kubectl rollout restart sts a-mycat-sts
kubectl apply -f og-old-a.yaml
kubectl rollout restart sts a-mycat-sts
