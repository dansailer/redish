#!/bin/bash
docker pull redis:latest
docker run --name my-redis -d -p 6379:6379 redis:latest