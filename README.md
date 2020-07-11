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
> 
> -- Carter Klein

## Design (as of 07/05/2020)

This system aims to follow domain driven design as closely as possible. All business logic should be clearly contained within easy-to-understand bounded contexts, using uniquitous language, wherever possible.

**Bounded contexts** in this case means things that are consistent within a bondary. A classic example of this is that you have a typical user. From a sales context, relevant metrics about this user are cost of acquisition, hours used daily, etc. From a human resources context, relevant metrics about this user are age, record, etc. These (sales, HR) are two separate contexts, and should be handled as such within an application.

### Building blocks for a context (ex in parentheses: authentication)

**Context:** The overall bounded context. *(Authentication)*

**Language:** Actual language used within the codebase. An example from my past experience is when I was on a team handling something related to an aggregated pay file for clients. Across the prdoduct, I saw the language "payfile," "payFile," "paymentFile," etc... this led to endless frustration and a compeltely disjoint experience. We should strive to not replicate that. That was clearly an example of moving fast to move slow, and we do the opposite. *(login, createAccount, vendor, entity)*

**Entity:** An abstract concept that could have actual instances. *(Vendor, Entity)*

**Value Object** A value on its own that isn't an entity on its own, but is a part of it. 

**Aggregate:** An object comprised of multiple entities

**Service:** An operation an entity shouldn't be doing on its own *(Vendor adding, Entity adding, Vendor listing, Entity listing, etc...)*

**Events:** Capture something interesting that happened in your system that effects the state of your system *(Vendor added, Entity Added, Vendor not found, etc...)*

**Repository:** Sits between domain logic and actual storage or database *(Vendor respoitory, entity repsoitory)*
- **NOTE:** while these are two separate databases within Supertype, they need not always be two separate repositories. This is an abstraction to better organize our minds.

## Architecture (as of 07/05/2020)

This system aims to implement a hexagonal architecture. Simply put, this means that there are layers (framework, application, domain, and core domain) with dependencies only pointing inwards. We want things to be as interchangeable as possible, especially for a system like Supertype. While we're only initially interacting with DynamoDB, we may eventually seek to communicate with different types of databases, or store data in a more cost-efficient and temporary way, such as on IPFS. By creating a generic `Repository` interface, for example, it becomes trivial to replace the source of that `Repository` as we see fit, whether that's DynamoDB, MongoDB, or IPFS - as long as the overall implementation remains the same.

## Troubleshooting 

- Ensure your AWS Security Tokens are set! They should be saved on your machine, and you configure them by running `aws configure` (assuming you have the AWS CLI set up)