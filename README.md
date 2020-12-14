# Supertype

[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://github.com/super-type/supertype)

This repo is the backbone of Supertype's device-agnostic data network. Almost all backend functionality lives within this repo, as our service is not yet large enough to justify splitting into multiple repositories. This monorepo acts as the central server for our system, and all data ingress and egress moves through the HTTP rotuer. Current contexts supported are:

* Authenticating: How vendors and users register and authenticate and authorize communication with Supertype.
* Consuming: How vendors can receive data from Supertype's data network.
* Producing: How vendors can produce data to Supertype's data netowrk.
* Vendor Management: Fetching vendor information regarding account infromation, payments, etc - primarily used for billing and vendor dashboard purposes
* Attribute Management: The constantly-solidifying mechanism of how vendors produce and consume data to and from the Supertype data network.

## Terms and definitions
**vendor:** a smart home device maker. These are the companies actually using Supertype.

**user:** a vendor's end user. The individual smart home owner/operator who is using vendor devices.

**attribute:** a specific entity that vendor devices observe. For example, the status of garage lights. This is currently a URI-based scheme (to be documented in more detail within the WIP Attribute Catalog), with the former attribute described as `/garage/lights/status`. 

**observation:** a specific measurement taken by one vendor device for a specific attribute. For example, a smart lightbulb chanigng from "off" to "on." 

**produce:** when a vendor device adds an observation to Supertype's data network.

**consume:** when a vendor device gets an observation from Supetype's data network.

## What is Supertype?

Supertype is a data network that allows smart home device makers to harness their ambient data that hubs like Google Home or Alexa aren’t designed to capture, and dev tools like IFTTT or Wink make the end user pay for. Our goal is to create a truly ambient smart home. Any individual can update their smart thermostat, regardless of which brand, and that action is immediately and automatically propagated to all other smart devices in the home.

## How Supertype works

Supertype is a managed data transfer network, with a growing number of touchpoints for vendors of all types to interact with our system. With simplicity in mind, Supertype takes form as an API client library with simple onramps to our data network. Vendors can produce data to and consume data from the Supertype data network.

Unlike other companies, Supertype does not make assumptions on behalf of our vendors. We do not focus routines or predefined actions, our focus lies in getting as much data as possible into vendors' hands.

## Vendor and user login flow

![Vendor Login](internal/images/vendor-login.png?raw=true "Vendor Login")
![User Login](internal/images/user-login.png?raw=true "User Login")

### Current Support

Supertype curently supports the following onramps:
- HTTP

Supertype currently supports the following API client libraries:
- Golang

## Encryption

Supertype uses a custom encryption protocol to guarantee efficient end-to-end encryption across all vendor devices. Loosely inspired by the [JEDI Encryption Protocol](https://arxiv.org/abs/1905.13369), Supertype's protocol aims to reduce the number of built-in features with the benefit of only relying on AES encryption, one of the most common forms of symmetric encryption.

The high-level overview of Supertype's encryption protocol is as follows:

1. When a user logs into Supertype through a vendor device, the user's information along with their unique AES encryption key is returned to the vendor.
2. The vendor can then use this AES encryption key to uniquely encrypt data for this user before uploading it to the Supertype data network. 

A more detailed overview can be found below:

![Encryption](internal/images/encryption.png?raw=true "Encryption")

## Branching Rules

1. All pull requests must contain 1 squashed commit, rebased off the current master branch
2. Branches and commit messages should have the following prefixes:

    * FEAT: For new functionality
    * BUG: For bug fixes
    * DEBT: For cleaning up tech debt
    * HOTFIX: For a small, ticket-less, direct-to-master hotfix

## Local setup
###  Prerequisites
1. Ensure account created on AWS and AWS is configured locally using AWS CLI
2. Golang, Redis, git installed on local machine 

### Build and Run
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

## Troubleshooting 

- Ensure your AWS Security Tokens are set! They should be saved on your machine, and you configure them by running `aws configure` (assuming you have the AWS CLI set up)
- Ensure you've set your environment variables (JWT, etc...)
- When testing with Postman, ensure you're using a recent JWT by first hitting the `/loginVendor` endpoint with valid credentials and inspecting the response
