ip=$(kubectl get pod -owide -n test | grep "d-masters-0" | awk '{print $6}')
# echo $ip
# # echo 8v16G-by-pod: >> test-result.txt
# for i in {8,16}
# do
#     echo thread $i >> test-result.txt
#     sysbench --db_driver=pgsql --pgsql-host=$ip \
#     --pgsql-port=5432 --pgsql-user=gaussdb --pgsql-password=Enmo@123 \
#     --pgsql-db=postgres  --threads=$i --rate=0 --time=120 --report-interval=5 oltp_read_only \
#     --tables=20 --table_size=100000 \
#     --skip_trx=on run | grep -E 'transactions|avg:' >> test-result.txt
#     sleep 5
# done

sysbench --db_driver=pgsql --pgsql-host=$ip \
--pgsql-port=5432 --pgsql-user=gaussdb --pgsql-password=Enmo@123 \
--pgsql-db=postgres  --threads=1 --rate=0 --time=10 --report-interval=5 oltp_read_only \
--tables=10 --table_size=10000 \
--skip_trx=on run