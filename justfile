default:
  @just --choose

# go build (local)
build:
  go build -v -o Syringe

# clean
clean:
  go clean
  rm *.exe
  rm go_build*

list-projects:
  ./Syringe list-projects -m

# go run-phylum (local)
run-phylum:
  ./Syringe run-phylum -m

# go run with pidfile as flag
run-phylum-pids:
  ./Syringe run-phylum --pidFilename pids.txt

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
