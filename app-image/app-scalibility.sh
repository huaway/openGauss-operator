run() {
    echo "cpu " $1 " memory " $2 >> scalibility.txt
    echo "---------------------" >> scalibility.txt
    if [ "$cpu" == "8" ]; then
        nums="1"
    else
        nums="6 7 8"
    fi
    for num in $nums
    do
        # 创建集群
        ./change $cpu $memory $num
        kubectl apply -f ../example/opengauss-test.yaml
        sleep 60

        echo "num " $num | tee -a scalibility.txt
        echo "::::::::::::::::" | tee -a scalibility.txt
        ip=$(kubectl get svc -n test | grep "shardingsphere" | awk '{print $3}')
        if [ "$cpu" == "1" ]; then
            thread=`expr $num \* 1`
        elif [ "$cpu" == "2" ]; then
            thread=`expr $num \* 2`
        elif [ "$cpu" == "4" ]; then
            thread=`expr $num \* 4`
        else
            thread=`expr $num \* 8`
        fi
        echo "thread " $thread | tee -a scalibility.txt

        sysbench --db_driver=pgsql --pgsql-host=$ip \
        --pgsql-port=5432 --pgsql-user=root --pgsql-password=root \
        --pgsql-db=postgres  --threads=$thread --rate=0 --time=120 --report-interval=5 oltp_read_only \
        --tables=10 --table_size=10000 \
        --skip_trx=on run | grep -A 30 "SQL statistics:" >> scalibility.txt
        echo ":::::::::" | tee -a scalibility.txt
        
        kubectl delete og d
        sleep 30
    done
    echo "---------------------" | tee -a scalibility.txt
}

for cpu in 8
do
    if [ "$cpu" == "1" ]; then
        memory=2
    elif [ "$cpu" == "2" ]; then
        memory=4
    elif [ "$cpu" == "4" ]; then
        memory=8
    else
        memory=16
    fi

    run $cpu $memory
done