ip=$(kubectl get pod -owide -n test | grep "d-masters-0" | awk '{print $6}')
psql -d postgres -U gaussdb -W -h $ip