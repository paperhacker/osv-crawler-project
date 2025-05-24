# OSV Crawler Project

A Go-based web crawler that runs in AWS ECS Fargate, pulls URLs from S3, scrapes them using Colly, and stores results back to S3.

## Build

```bash
make build
```

## Run Locally

```bash
make run
```

## Deploy Infrastructure

```bash
cd terraform
terraform init
terraform apply
```

## Trigger Crawler

Upload `target.json` to S3:

```bash
aws s3 cp target.json s3://crawler-input-bucket/target.json
```
