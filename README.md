
NATS-Consumers-78214263
=======================

This repo contains code for [my answer][so:answer] to question ["How are Consumers implemented in NATS"][so:question] posted on [StackOverflow][so].


Prerequisites
-------------

* docker
* docker compose
* git


Run
---

1. Check out the code
2. Run `docker compose build`
3. Run `docker compose up [-d]`

Then, watch the logs of service "consumer".

[so]: https://www.stackoverflow.com
[so:question]: https://stackoverflow.com/questions/78214263
[so:answer]: https://stackoverflow.com/a/78214960/1296707