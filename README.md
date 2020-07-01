# Supertype

[![Go Report Card](https://goreportcard.com/badge/github.com/golang-standards/project-layout?style=flat-square)](https://github.com/super-type/supertype)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://github.com/super-type/supertype)
[![Release](https://img.shields.io/github/release/golang-standards/project-layout.svg?style=flat-square)](https://github.com/super-type/supertype)

This is where we will keep our backend functionality in the form of a smart monolith, for the time being. The microservices-first path is wrought with peril, and we wish to make our lives as easy as possible. Which, at a startup, is a challenge in and of itself. Many of the best minds in the software architecture space ([Martin Fowler](https://martinfowler.com/bliki/MonolithFirst.html), [Atlassian Blog](https://www.atlassian.com/continuous-delivery/microservices/building-microservices)) argue that starting with a monolith is the simplest way to get code off the ground.

Building an architecture is inherently complicated, and we're inherently bad at that. As a startup, requirements, needs, and client requests are constantly changing, and we want to start with an architecture that's best suited for that. While we have put in a signficant amount of effort into determing what our end-state, microservices architecture may look like, we don't have a crystal ball.

Therefore, we will start with a monolith. Once we have a specific service within our monolith that warrants a microservices, we'll spin it out to become our first microservice. Until then, monolith it is.
