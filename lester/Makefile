docker-rabbitmq:
	docker run -d --rm --name rabbitmq \
	-p 5672:5672 -p 15672:15672 \
	-e RABBITMQ_DEFAULT_USER=user \
	-e RABBITMQ_DEFAULT_PASS=pass \
	rabbitmq:3-management

docker-lester: docker-rabbitmq
	docker build -t lester .
	docker run --rm --name lester -p 50051:50051 \
		-e RABBITMQ_HOST=10.35.168.89 \
		-e RABBITMQ_PORT=5672 \
		-e RABBITMQ_DEFAULT_USER=user \
		-e RABBITMQ_DEFAULT_PASS=pass \
		lester
