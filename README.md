# BORALabs DAO API

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Installation](#installation)

## Introduction

This service is developed in Go and runs a web server using Gin-Gonic.

It periodically collects blockchain event logs to update the information of created proposals and votes.

## Features

- Provide RESTFul API to the frontend through Gin-Gonic
- Periodically collect blockchain event logs
- Update proposal and vote information

## Installation

Local installation is done by building the container through docker-compose. You need to configure `app.yaml`.

Step-by-step instructions:

1. Clone the repository
    ```bash
    git clone https://github.com/boraecosystem/boralabs-dao-api.git
    cd boralabs-dao-api
    ```

2. Configure config/`app.yaml`
    ```yaml
    # Example configuration
      env: local
      fromBlock: 0
      isRewind: false
      debug: true
      rpcEndpoint: ""
      mongo:
        host: boralabs-dao-db
        port: 27017
        user: root
        pass: 1234
      slack:
       webhook_url: ""
      daoAddress: ""
      governorAddress: ""
    ```

3. Build and run the container
    ```bash
    docker-compose up -d
    ```
