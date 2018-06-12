.PHONY: backup

resetdb:
	psql -h localhost -U postgres -c "drop database todo"
	psql -h localhost -U postgres -c "create database todo owner todo"

run:
	# go get github.com/codegangsta/gin
	gin --port 3500 --appPort 4000

build:
	docker build --tag yegorlitvinov/todo-back:latest .

SERVER=todo@195.201.27.44
ROOT_SERVER=root@195.201.27.44
HOME=/home/todo
deploy-app:
	scp docker-compose.yml $(SERVER):$(HOME)
	ssh $(SERVER) 'cd $(HOME) && docker-compose pull && docker-compose up -d'

deploy-nginx:
	scp todo_nginx.conf $(ROOT_SERVER):/etc/nginx/sites-enabled/
	ssh $(ROOT_SERVER) 'nginx -t'
	ssh $(ROOT_SERVER) 'service nginx restart'
	# certbot certonly -d *.tvgun.ga --server https://acme-v02.api.letsencrypt.org/directory --manual
	# certbot renew

BACKUP_FILE=backup/todo.dump
backup:
	mkdir -p backup
	ssh $(SERVER) 'mkdir -p ./backup'
	ssh $(SERVER) "docker-compose exec -T -u postgres postgres pg_dump -Fc todo > $(BACKUP_FILE)"
	rsync -aP --delete -e ssh $(SERVER):$(BACKUP_FILE) `pwd`/$(BACKUP_FILE)

restore:
	pg_restore -d todo -U todo -h localhost -Fc $(BACKUP_FILE)
