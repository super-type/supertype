# Supertype

[![Go Report Card](https://goreportcard.com/badge/github.com/golang-standards/project-layout?style=flat-square)](https://github.com/super-type/supertype)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://github.com/super-type/supertype)
[![Release](https://img.shields.io/github/release/golang-standards/project-layout.svg?style=flat-square)](https://github.com/super-type/supertype)

This is where we will keep our backend functionality in the form of a smart monolith, for the time being. The microservices-first path is wrought with peril, and we wish to make our lives as easy as possible. Which, at a startup, is a challenge in and of itself. Many of the best minds in the software architecture space ([Martin Fowler](https://martinfowler.com/bliki/MonolithFirst.html), [Atlassian Blog](https://www.atlassian.com/continuous-delivery/microservices/building-microservices)) argue that starting with a monolith is the simplest way to get code off the ground.

Building an architecture is inherently complicated, and we're inherently bad at that. As a startup, requirements, needs, and client requests are constantly changing, and we want to start with an architecture that's best suited for that. While we have put in a signficant amount of effort into determing what our end-state, microservices architecture may look like, we don't have a crystal ball.

Therefore, we will start with a monolith. Once we have a specific service within our monolith that warrants a microservices, we'll spin it out to become our first microservice. Until then, monolith it is.

> README hygeine (as of 07/05/2020)
>
> Something that I have found personall frustrating in my limited experience with building software is a strikign disconnect between initial architecture decisions, and a product's current state. Perhaps it is from my tenure at a decidedly non-tech insurance firm, and a large bank feigning to be "tech," but I've found that time after time, a prodcut's inception is completely decoupled from its developemnt. This leaves it up to tech leads, developers, and product managers (who may have joined this product months if not years into its development) to learn the history of the project, and why it currently looks the way it does. While there will often be an endless supply of source material in the way of architecture diagrams, meeting notes, and the like - this does not paint the full picture of the project and leaves developers on the product from its inception with a lasting advantage over those newer to the project. Therefore, it's of utmost importance at Supertype that we enact one of our guiding principles, and **move slow to go fast.** In this case, I'm talking documentation. I believe that documentation should not just provide the requisite few badges, a brief description of the project, directions on how to run/test the product locally and in QA, and an overview of deployment steps. The documentation should cover the product's history. A manager of mine would emphasize that as software engineers, what we deliver is not code, but a sequence of git commits that describe the life of a product. Needless to say, he was a stickler (in the best possible connotation of the term) on git hygeiene.
>
> I believe that taking this a step further and not just providing a history of decisions through git commits, but a thorough, history of the product as a whole is crucial to its longevity and its ability to maintain or gain traction over time. I have seen teams spend months getting acclimated with a product. Digging through the codebase(s), reading git histories, taking part in many knowledge transfer sessions over many weeks. However, most questions about the codebase (often something along the lines of "why the *fuck* would A have done B back in 20CC?") would be answered once someone uncovered a decision made at the time which led to the code in qeustion, and oftentimes code for the subsequent couple of weeks. Looking atomically at a piece of code in a years-old repository will often lead to hours of head-scratching that could be largely abated by a clear and explanatory product history outlining initial design/architecture decisions, significant changes made to the codebase, workarounds, and discussions about what fit where. At Supertype, there is not such thing as too long of a README - too *short* a README is when things get concerning...
>
> Update the README whenever you feel it's necessary. Be sure to update the "as of" addendum to each section, so that future developers can combine information from this README with a comprehensive git history for a comprehensive understanding of the product as a whole.

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
    "username": <USERNAME>,
    "password": <PASSWORD>
}
```

**/createvendor: (POST):** Generates a new vendor
- body:
```json
{
    "username": <USERNAME>,
    "password": <PASSWORD>,
    "firstName": <FIRST NAME>, // Optional
    "lastName": <LAST NAME> // Optional
    // TODO update with more as we create more...
}
```

**/produce: (POST):** Produces data for a specific Supertype type user from a specific vendor to the Supertype ecosystem. Also, for the time being, runs any additional necessary re-encryptions with new vendors to ensure each vendor is up to date.
- headers:
    - `Token` : `<JWT GENERATED ON LOGIN>`
- body:
```json
{
    "ciphertext": <CIPHERTEXT>,
    "capsule": <CAPSULE>,
    "attribute": <ATTRIBUTE>
}
```
- **NOTE** the ciphertext and capsule are generated from the `goImplement` (or any future implementations) package

**/consume: (POST):** Consumes data for a specific user from the Supertype ecosystem, regardless of which vendor produced it
- headers:
    - `Token` : `<JWT GENERATED ON LOGIN>`
- body:
```json
{
    "attribute": <ATTRIBUTE>
}
```

**/getVendorComparisonMetadata: (POST):** Returns collections of all vendors, as well as the currently logged-in vendor's connections, in order to determine if additional re-encryptions are necessary
- headers:
    - `Token` : `<JWT GENERATED ON LOGIN>`
- body:
```json
{
    "attribute": <ATTRIBUTE>,
    "supertypeID": <SUPERTYPE ID>,
    "pk": <VENDOR PUBLIC KEY>
}
```

**/addReencryptionKeys: (POST):** Runs re-encryption from current vendor to any newly-created vendors, if necessary
- headers:
    - `Token` : `<JWT GENERATED ON LOGIN>`
- body:
```json
{
    "connections": <CONNECTIONS>,
    "pk": <VENDOR PUBLIC KEY>
}
```

## Troubleshooting 

- Ensure your AWS Security Tokens are set! They should be saved on your machine, and you configure them by running `aws configure` (assuming you have the AWS CLI set up)
- Ensure you've set your environment variables (JWT, etc...)

## TODO 
- go into architecture
- go into database structure (how we have a global secondary index on pk in vendor because we use it for adding re-encryption keys)