# ThreatWatch

This repository hosts a cloud-native Threat Intelligence Service developed in Go. The service is built using a microservices architecture, enabling scalability, modularity, and simplified maintenance. Please note that this project is intended as a proof of concept and may not be suitable for production use. It served me as a practical learning resource for understanding the implementation of microservices and exploring cloud-native architecture concepts.

## Architecture

The Cloud Native Threat Intelligence Service follows a microservices architecture, dividing the functionality into smaller, loosely coupled services. Each microservice handles a specific aspect of the threat intelligence process, such as data gathering, analysis, and retrieval. The communication between microservices is typically performed through APIs or messaging protocols.

The repository includes the following microservices:

- **Auth**: Used for JWT authentication
- **Users**: Service to manage user accounts
- **IPv4**: Report malicious IPv4 addresses and get reports
- **Logging**: Store application logs
- **Healthcheck**: Healthcheck service

![TW_arch](https://github.com/dev-lu/threatwatch/assets/44299200/e689299c-3dae-44fe-80cb-f5081dcc8503)


## Getting Started

To try this project, follow the steps below.

1. Install Docker and Docker Compose

2. Clone the repository:
```shell
git clone https://github.com/dev-lu/threatwatch.git
```

3. Start the Docker containers
```shell
docker-compose up --build
```
