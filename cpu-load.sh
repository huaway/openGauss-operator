# 当以不同速率发送请求时，求出CPU利用率
run() {
    sysbench --db_driver=pgsql --pgsql-host=$1 \
	--pgsql-port=5432 --pgsql-user=root --pgsql-password=root \
	--pgsql-db=postgres  --threads=$2 --time=$3 --report-interval=5 oltp_read_write \
	--tables=20 --table_size=100000 \
	--skip_trx=on &\
    
    sleep 30

    go run processor/cpu_load.go
}