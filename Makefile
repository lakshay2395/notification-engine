docker.start:
	docker-compose down
	docker volume prune -f
	docker rmi -f manager-service:latest
	docker build -f ./manager-service/Dockerfile -t manager-service:latest ./manager-service
	docker rmi -f worker-service:latest
	docker build -f ./worker-service/Dockerfile -t worker-service:latest ./worker-service
	docker-compose up

docker.stop:
	docker-compose down
	docker volume prune -f
	docker rmi -f manager-service:latest
	docker rmi -f worker-service:latest

start: docker.start