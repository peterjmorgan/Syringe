default:
  @just --choose

# go build (local)
build:
  go build -v -o Syringe

# go run (local)
run:
  ./Syringe run-phylum

# docker build image
docker-build:
  @docker build -t syringe .

# docker run run-phylum
docker-run-phylum:
  docker run -e PHYLUM_API_KEY -e GITLAB_TOKEN syringe run-phylum

# docker run list-projects
docker-run-list-projects:
  docker run -e PHYLUM_API_KEY -e GITLAB_TOKEN syringe list-projects

# docker run run-phylum -d (debug)
docker-run-phylum-debug:
  docker run -e PHYLUM_API_KEY -e GITLAB_TOKEN syringe run-phylum -d

# docker run syringe image interactive shell
docker-run-shell:
  docker run -e PHYLUM_API_KEY -e GITLAB_TOKEN -it --entrypoint bash syringe
