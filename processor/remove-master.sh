kubectl delete -f og-b.yaml
kubectl rollout restart sts a-mycat-sts
kubectl apply -f og-a-old.yaml
kubectl rollout restart sts a-mycat-sts