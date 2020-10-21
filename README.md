# Supertype

[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://github.com/super-type/supertype)

This is where we will keep our backend functionality in the form of a smart monolith, for the time being. The microservices-first path is wrought with peril, and we wish to make our lives as easy as possible. Which, at a startup, is a challenge in and of itself. Many of the best minds in the software architecture space ([Martin Fowler](https://martinfowler.com/bliki/MonolithFirst.html), [Atlassian Blog](https://www.atlassian.com/continuous-delivery/microservices/building-microservices)) argue that starting with a monolith is the simplest way to get code off the ground.

Building an architecture is inherently complicated, and we're inherently bad at that. As a startup, requirements, needs, and client requests are constantly changing, and we want to start with an architecture that's best suited for that. While we have put in a signficant amount of effort into determing what our end-state, microservices architecture may look like, we don't have a crystal ball.

Therefore, we will start with a monolith. Once we have a specific service within our monolith that warrants a microservices, we'll spin it out to become our first microservice. Until then, monolith it is.

## Architecture

![Architecture](internal/images/architecture-9-10.png?raw=true "Architecture")

## Build & Run Locally
1. `git clone https://github.com/super-type/supertype`
2. `cd supertype`
3. `make run`

## API Endpoints

**/healthcheck: (GET):** A simple healthcheck to ensure you're running everything properly

**/loginVendor: (POST):** Logs in a pre-existing vendor to the Supertype ecosystem
- body:
```json
{
    "username": "<USERNAME>",
    "password": "<PASSWORD>"
}
```

**/createvendor: (POST):** Generates a new vendor
- body:
```json
{
    "username": "<USERNAME>",
    "password": "<PASSWORD>",
    "firstName": "<FIRST NAME>",
    "lastName": "<LAST NAME>"
}
```

**/produce: (POST):** Produces data for a specific Supertype type user from a specific vendor to the Supertype ecosystem. Also, for the time being, runs any additional necessary re-encryptions with new vendors to ensure each vendor is up to date.
- headers:
    - `Token` : `<JWT GENERATED ON LOGIN>`
- body:
```json
{
    "ciphertext": "<CIPHERTEXT>",
    "attribute": "<ATTRIBUTE>"
}
```
- **NOTE** the ciphertext is generated from the `goImplement` (or any future implementations) package

**/consume: (POST):** Consumes data for a specific user from the Supertype ecosystem, regardless of which vendor produced it
- headers:
    - `Token` : `<JWT GENERATED ON LOGIN>`
- body:
```json
{
    "attribute": "<ATTRIBUTE>"
}
```

**/consume: (WebSocket):** Creates a WebSocket connection between vendor and server and subscribes the vendor to realtime updates on specified attributes

## Troubleshooting 

- Ensure your AWS Security Tokens are set! They should be saved on your machine, and you configure them by running `aws configure` (assuming you have the AWS CLI set up)
- Ensure you've set your environment variables (JWT, etc...)
- When testing with Postman, ensure you're using a recent JWT by first hitting the `/loginVendor` endpoint with valid credentials and inspecting the response
