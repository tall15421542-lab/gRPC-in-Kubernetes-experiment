BLUE=\033[0;32m
NO_COLOR=\033[0m

build-image: 
	@printf 'run [${BLUE}eval $$(minikube -p minikube docker-env)${NO_COLOR}] command first\n'
	docker build -t greeter-server .
	docker build -t greeter-client .
