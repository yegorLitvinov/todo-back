resetdb:
	psql -h localhost -U postgres -c "drop database todo"
	psql -h localhost -U postgres -c "create database todo owner todo"

run:
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
