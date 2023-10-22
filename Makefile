.SILENT:

infra:
	sudo docker-compose up -d --build
infra-stop:
	sudo docker-compose down && sudo docker image prune