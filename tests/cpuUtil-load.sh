# This file measure relationship of cpuUtil and load

ip=$(kubectl get svc -n test | grep "shardingsphere" | awk '{print $3}')

# echo "1v2G:" | tee -a relation-cpu.txt
# echo "1v2G:" | tee -a relation-load.txt
# for thread in {4}
# do
# echo "thread " $thread ": " | tee -a relation-cpu.txt
# echo "thread " $thread ": " | tee -a relation-load.txt
# for rate in {15,20}
# do
# echo "rate " $rate ": " | tee -a relation-cpu.txt
# echo "rate " $rate ": " | tee -a relation-load.txt
# ./detect 300 >> relation-cpu.txt &

# sysbench --db_driver=pgsql --pgsql-host=$ip \
# --pgsql-port=5432 --pgsql-user=root --pgsql-password=root \
# --pgsql-db=postgres  --threads=$thread --rate=$rate --time=300 --report-interval=5 oltp_read_write \
# --tables=10 --table_size=10000 \
# --skip_trx=on \
# run \
# >> relation-load.txt
# sleep 10
# done
# done

# kubectl delete og d
# sleep 20
# kubectl apply -f ../example/opengauss-new.yaml
# sleep 30
# kubectl delete po d-sts-0
# sleep 60
# for thread in {1,2,4}
# do
# echo "thread " $thread ": " | tee -a relation-cpu-mid.txt
# echo "thread " $thread ": " | tee -a relation-load-mid.txt
# for rate in {11,15,21,23,25,31,35}
# do
# echo "rate " $rate ": " | tee -a relation-cpu-mid.txt
# echo "rate " $rate ": " | tee -a relation-load-mid.txt
# ./detect 240 >> relation-cpu-mid.txt &

# sysbench --db_driver=pgsql --pgsql-host=$ip \
# --pgsql-port=5432 --pgsql-user=root --pgsql-password=root \
# --pgsql-db=postgres  --threads=$thread --rate=$rate --time=240 --report-interval=5 oltp_read_write \
# --tables=10 --table_size=10000 \
# --skip_trx=on \
# run \
# >> relation-load-mid.txt
# sleep 10
# done
# done

for thread in 4
do
echo "thread " $thread ": " | tee -a relation-cpu-large.txt
echo "thread " $thread ": " | tee -a relation-load-large.txt
for rate in {10,15,20,25,30,35,40,45,50}
do
echo "rate " $rate ": " | tee -a relation-cpu-large.txt
echo "rate " $rate ": " | tee -a relation-load-large.txt
./detect 240 >> relation-cpu-large.txt &

sysbench --db_driver=pgsql --pgsql-host=$ip \
--pgsql-port=5432 --pgsql-user=root --pgsql-password=root \
--pgsql-db=postgres  --threads=$thread --rate=$rate --time=240 --report-interval=5 oltp_read_write \
--tables=10 --table_size=10000 \
--skip_trx=on \
run \
>> relation-load-large.txt
sleep 10
done
done