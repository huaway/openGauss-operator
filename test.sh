unit(){
	sysbench --db_driver=pgsql --pgsql-host=$1 \
	--pgsql-port=5432 --pgsql-user=root --pgsql-password=root \
	--pgsql-db=postgres  --threads=$2 --time=$3 --report-interval=5 oltp_read_write \
	--tables=20 --table_size=100000 \
	--skip_trx=on \
	run
}

run(){
	for ((; i<=$1; i++))
	do
		run $ip 1 60
		run $ip 
}
ip=$(kubectl get svc -n test | grep "shardingsphere" | awk '{print $3}')
run $ip 1