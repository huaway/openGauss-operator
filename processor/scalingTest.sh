go run ./test-1static.go &
sleep 300

go run ./test-2static.go &
sleep 600

go run ./test-3static.go &
sleep 300

go run ./test-2static.go &
sleep 600

go run ./test-1static.go &
