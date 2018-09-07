# perm

This Permissions service ("perm") provides authorization features for the Cloud
Foundry Platform. It answers various question forms of what particular
identities are allowed to do. It works out the answers to these questions based
on the roles assigned to users and the roles assigned to the groups they are a
member of.

Even though the service was originally created to add authorization features to
Cloud Controller, other components in the system are looking to migrate to
storing their authorization rules in Perm.

### Installation

To fetch all **source code**, including the **Go client library**:

```bash
go get -u code.cloudfoundry.org/perm
```

To fetch and install the **server's CLI**:

```bash
go get -u code.cloudfoundry.org/perm/cmd/perm
```

To fetch and install the **monitor's CLI**:

```bash
go get -u code.cloudfoundry.org/perm/cmd/perm-monitor
```

### Running the Tests

Assuming you have the Perm source code in your $GOPATH:

```
go install code.cloudfoundry.org/perm/vendor/github.com/onsi/ginkgo/ginkgo
ginkgo -r -race -p -randomizeAllSpecs -randomizeSuites
```

### Running the Perm Server

First, make sure that you have the CLI installed:

```bash
go get -u code.cloudfoundry.org/perm
go install code.cloudfoundry.org/perm/cmd/perm
```

To use an in-memory data store, e.g., for testing purposes:

```bash
perm serve --tls-cert <path> --tls-key <path> --db-driver in-memory
```

To use mysql:
```bash
perm migrate up --db-driver mysql --db-host <host> --db-port <port> --db-username <username> --db-password <password>
perm serve --tls-cert <path> --tls-key <path> --db-driver mysql --db-host <host> --db-port <port> --db-username <username> --db-password <password>
```

### Running the Perm Monitor

The monitor is a small app that repeats the same basic workflow every interval, generating traffic and tracking some client-side metrics.

Make sure that you have the monitor's CLI:

```bash
go get -u code.cloudfoundry.org/perm
go install code.cloudfoundry.org/perm/cmd/perm-monitor
```

Make sure that you have a [statsd](https://github.com/etsy/statsd) daemon, e.g., with docker:

```bash
docker run -d -p 8125:8125 --name statsd hopsoft/graphite-statsd
```

Then, start the monitor:

```bash
perm-monitor --perm-tls-ca <path>
```

### In Popular Culture

This repository is complemented by 2 other repositories.

* [perm-release](https://github.com/cloudfoundry-incubator/perm-release)

  This is the BOSH release for deploying the `perm` service.

* [perm-rb](https://github.com/cloudfoundry-incubator/perm-rb)

  This is the Ruby library for interacting with `perm`. It is used by Cloud
  Controller to perform administration and checking of permissions.

For more information, check out our page on [Repository Structure](https://github.com/cloudfoundry-incubator/perm/wiki/Repository-structure).

### Usage

Not yet, please.
