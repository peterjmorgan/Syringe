default:
  @just --list

dbuild:
  docker build -t syringe .

drun:
  docker run -e PHYLUM_API_KEY -e GITLAB_TOKEN syringe run-phylum

drundebug:
  docker run -e PHYLUM_API_KEY -e GITLAB_TOKEN syringe run-phylum -d

dshell:
  docker run -e PHYLUM_API_KEY -e GITLAB_TOKEN -it --entrypoint bash syringe
