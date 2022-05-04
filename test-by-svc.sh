# smallx="small"
# midx="mid"
# largex="large"
# for t in small
# do
#     if [ "$t" = "$midx" ];then
#         echo 2v4G-by-shardingsphere: >> test-result.txt
#     fi
#     if [ "$t" = "$smallx" ];then
#         echo 1v2G-by-shardingsphere: >> test-result.txt
#     fi
#     if [ "$t" = "$largex" ];then
#         echo 4v8G-by-shardingsphere: >> test-result.txt
#     fi

#     kubectl delete og d
#     sleep 10
#     kubectl apply -f example/opengauss-$t-1.yaml
#     sleep 10
#     ip=$(kubectl get svc -n test | grep "shardingsphere" | awk '{print $3}')
#     for i in {1,2,4,8}
#     do
#         echo thread $i >> test-result.txt
#         sysbench --db_driver=pgsql --pgsql-host=$ip \
#         --pgsql-port=5432 --pgsql-user=root --pgsql-password=root \
#         --pgsql-db=postgres  --threads=$i --rate=0 --time=120 --report-interval=5 oltp_read_only \
#         --tables=20 --table_size=100000 \
#         --skip_trx=on \
#         run | grep -E 'transactions|avg:' >> test-result.txt
#         sleep 10
#     done
# done

# kubectl delete og d
# sleep 10
# kubectl apply -f example/opengauss-vlarge-1.yaml
# sleep 10
ip=$(kubectl get svc -n test | grep "shardingsphere" | awk '{print $3}')
sysbench --db_driver=pgsql --pgsql-host=$ip \
--pgsql-port=5432 --pgsql-user=root --pgsql-password=root \
--pgsql-db=postgres  --threads=5 --rate=0 --time=500 --report-interval=5 oltp_read_write \
--tables=10 --table_size=10000 \
--skip_trx=on \
run