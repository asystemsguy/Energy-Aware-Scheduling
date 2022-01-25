build: ## Build the container
	docker build -t khv129/app2 .

run:
	docker run -d -p 8090:8090 --name="app2" khv129/app2:latest


	curl -H 'X-My-Header: 123' -X GET http://0.0.0.0:8090/