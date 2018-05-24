resetdb:
	psql -h localhost -U postgres -c "drop database todo"
	psql -h localhost -U postgres -c "create database todo owner todo"

run:
	gin --port 3500 --appPort 4000

install-req:
	cat req.txt | go get
	go get github.com/codegangsta/gin
	