# bay
Docker jails for for running untrusted code.

Bay is a golang library to expose a simple API to run untrusted code in a Docker container. Obviously, this means it is Linux only and all the security concerns around running untrusted code in a Linux contianer still apply.

# Security
Can't stress this enough. *YOU* need to make sure you follow the best practices for running Linux containers. This library does not magically lock down your containers in a complete safe and isolated manner. Bay is only as strong as Docker. Checkout these resources regarding Docker security.

- The [Docker Security](https://docs.docker.com/articles/security/) article from Docker.io.
- The [LXC, Docker, Security](http://www.slideshare.net/jpetazzo/linux-containers-lxc-docker-and-security) slides from Jérôme Petazzoni.
- The series of Docker security articles from Daniel J. Walsh ([one](http://opensource.com/business/14/7/docker-security-selinux), [two](http://opensource.com/business/14/9/security-for-docker)). 
- The [Linux Audit](http://linux-audit.com/docker-security-best-practices-for-your-vessel-and-containers/) for some additional best practices.

## Install
`go get github.com/Vluxe/bay`

## Docs

## Overview/How to Use

## TODOs

things...

## License

Bay is licensed under Apache v2.

## Contact

...
