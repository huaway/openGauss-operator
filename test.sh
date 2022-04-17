unit(){
	sysbench --db_driver=pgsql --pgsql-host=$1 \
	--pgsql-port=5432 --pgsql-user=root --pgsql-password=root \
	--pgsql-db=postgres  --threads=$2 --time=$3 --report-interval=5 oltp_read_write \
	--tables=20 --table_size=100000 \
	--skip_trx=on \
	run
}

run(){
	go run processor/example.go
	echo start scaler
	for ((; i<=$1; i++))
	do
		run $ip 1 60
		run $ip 2 60
		# run $
	done
}

# 启动两个进程，load generator & cpu load calculator
# generator 启动| 60s        sample 启动|  120s    sample停止 |  120s   generator停止 |  60s   开始下一轮|
# 				---------------------------------------------------------------------------------------
point() {
	sysbench --db_driver=pgsql --pgsql-host=$1 \
	--pgsql-port=5432 --pgsql-user=root --pgsql-password=root \
	--pgsql-db=postgres  --threads=$2 --time=$3 --report-interval=5 oltp_read_write \
	--tables=20 --table_size=100000 \
	--skip_trx=on \
	run > /tmp/sysbench &

	sleep 60

	go run processor/cpu_load.go $4

	sleep 180
}

line() {
	for i in {1..6}
	do
		echo ------start drawing point whose thread is $i
		point $1 $i 300 $2
	done
}

configure(){
	for i in {2..5}
	do
		echo ---start drawing line with $i replica
		kubectl delete og d
		sleep 60
		kubectl apply -f example/opengauss-$1-$i.yaml
		sleep 60
		ip=$(kubectl get svc -n test | grep "shardingsphere" | awk '{print $3}')
		line $ip $1
	done
}

for t in small
do
	echo start drawing lines about $t replica !!!
	configure $t
done

#启动10个线程，阈值法只有4v8G这一种规格
kubectl delete og d
sleep 10
kubectl apply -f opengauss/opengauss-large-1.yaml
sleep 5
go run processor