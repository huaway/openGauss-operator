ip=$(kubectl get pod -owide -n test | grep "d-replicas-small-0" | awk '{print $6}')
psql -d postgres -U gaussdb -W -h $ip