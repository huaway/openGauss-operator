ip=$(kubectl get svc -n test | grep "shardingsphere" | awk '{print $3}')
psql -d postgres -U root -W -h $ip