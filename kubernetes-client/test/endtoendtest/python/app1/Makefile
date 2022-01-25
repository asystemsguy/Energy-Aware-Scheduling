build: ## Build the container
	docker build -t khv129/app1 .

run:
	docker run -d -p 8090:8090 --name="app1" khv129/app1:latest


	curl -H 'X-My-Header: 123' -X GET http://0.0.0.0:8090/