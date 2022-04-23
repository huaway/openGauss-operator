# This file test cost and sla of tantative method
ip=$(kubectl get svc -n test | grep "shardingsphere" | awk '{print $3}')
for load in "1 5" "3 20" "1 5" "3 20" "4 30" "3 20" "1 5"
do
thread=$(echo $load | awk '{print $1}')
rate=$(echo $load | awk '{print $2}')
echo thread: $thread | tee -a tantative-cost.txt
echo rate: $rate | tee -a tantative-cost.txt
../processor/auto-scaler 300 >> tantative-cost.txt &

sysbench --db_driver=pgsql --pgsql-host=$ip \
--pgsql-port=5432 --pgsql-user=root --pgsql-password=root \
--pgsql-db=postgres  --threads=$thread --rate=$rate --time=300 --report-interval=5 oltp_read_write \
--tables=10 --table_size=10000 \
--skip_trx=on \
run \
>> tantative-sla.txt

sleep 10
done