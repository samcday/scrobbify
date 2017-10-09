sam-local: sh -c 'cd backend && BASE_URL=http://localhost:3000 aws-sam-local local start-api --docker-network scrobbify_default'
backend: sh -c 'cd backend && reflex -r "\.go$" make'
services: docker-compose -p scrobbify up --abort-on-container-exit
