# This file test cost and sla of threshold-based method
ip=$(kubectl get svc -n test | grep "shardingsphere" | awk '{print $3}')
../processor/auto-scaler 2600 >> threshold-cost.txt &
# for load in "1 4" "1 5" "1 8" "1 4" "1 5" "3 20" "3 21" "3 19" "3 19" "3 21" "1 5" "1 5" "1 5" "1 5" "1 5" "3 21" "3 19" "3 19" "3 21" "3 21" "4 31" "4 32" "4 47" "4 32" "4 32" "3 21" "3 19" "3 21" "3 20" "3 20" "1 5" "1 4" "1 5" "1 4" "1 4"
# for load in "4 40" "4 40" "4 40" "4 40"
for load in "1 5" "1 5" "1 5" "1 5" "1 5" "1 7" "1 9" "2 11" "2 20" "2 14" "2 20" "2 20" "2 20" "2 21" "2 25" "2 15" "2 13" "2 10" "1 8" "1 6" "1 5" "1 5" "1 5" "1 5" "1 5" "1 9" "2 21" "3 21" "4 34" "4 33" "4 32" "3 21" "2 19" "1 10" "1 5" "1 4" "1 4" "1 5" "1 5"
do
thread=$(echo $load | awk '{print $1}')
rate=$(echo $load | awk '{print $2}')
echo thread: $thread | tee -a threshold-cost.txt
echo rate: $rate | tee -a threshold-cost.txt

sysbench --db_driver=pgsql --pgsql-host=$ip \
--pgsql-port=5432 --pgsql-user=root --pgsql-password=root \
--pgsql-db=postgres  --threads=$thread --rate=$rate --time=60 --report-interval=5 oltp_read_write \
--tables=10 --table_size=10000 \
--skip_trx=on \
run \
>> threshold-sla.txt

sleep 5
done