.SILENT:

run:
	sudo docker-compose up -d --build
clean:
	sudo docker-compose down && sudo docker image prune