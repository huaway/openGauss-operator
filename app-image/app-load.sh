# This file measure relationship between pod specification and capacity

run() {
    for memory in $1
    do
        ./change $cpu $memory 1
        kubectl apply -f ../example/opengauss-test.yaml
        sleep 20
        echo "---------------------" >> result.txt
        echo "cpu " $cpu "||||| memory " $memory >> result.txt

        if [ "$2" = "4" ]; then
            threads="1 2 3 4 5 6 7 8"
        elif [ "$2" = "8" ]; then
            threads="1 2 3 4 5 6 7 8 9 10 11"
        else
            threads="1 2 3 4"
        fi

        # 直连
        echo "direct:::::::" >> result.txt
        ip=$(kubectl get pod -owide -n test | grep "d-replicas-small-0" | awk '{print $6}')
        for thread in $threads
        do
            echo "thread " $thread " :::" >> result.txt
            sysbench --db_driver=pgsql --pgsql-host=$ip \
            --pgsql-port=5432 --pgsql-user=gaussdb --pgsql-password=Enmo@123 \
            --pgsql-db=postgres  --threads=$thread --rate=0 --time=120 --report-interval=5 oltp_read_only \
            --tables=10 --table_size=10000 \
            --skip_trx=on run | grep -A 30 "SQL statistics:" >> result.txt
        done
        echo ":::::::::" >> result.txt

        # 通过中间件
        echo "middleware:::::::" >> result.txt
        ip=$(kubectl get svc -n test | grep "shardingsphere" | awk '{print $3}')
        for thread in $threads
        do
            echo "thread " $thread " :::" >> result.txt
            sysbench --db_driver=pgsql --pgsql-host=$ip \
            --pgsql-port=5432 --pgsql-user=root --pgsql-password=root \
            --pgsql-db=postgres  --threads=$thread --rate=0 --time=120 --report-interval=5 oltp_read_only \
            --tables=10 --table_size=10000 \
            --skip_trx=on run | grep -A 30 "SQL statistics:" >> result.txt
        done
        echo ":::::::::" >> result.txt
        
        echo "---------------------" >> result.txt
        kubectl delete og d
        sleep 20
    done
}

for cpu in 0.25 0.5 1 2 4 8 12 16
do
    # if [ "$cpu" = "0.25" ]; then
    #     run "0.5 1" 0.25
    # fi
    if [ "$cpu" = "0.5" ]; then
        run "1 2" 0.5
    fi
    # if [ "$cpu" = "1" ];then
    #     run "2 3 4" 1
    # fi
    # if [ "$cpu" = "2" ];then
    #     run "4 5 6 7 8" 2 
    # fi
    # if [ "$cpu" = "4" ];then
    #     run "8 9 10 11 12 13 14 15 16" 4 
    # fi
    # if [ "$cpu" = "8" ];then
    #     run "8 16 32" 8
    # fi
done