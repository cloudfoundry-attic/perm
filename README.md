# perm

## About

This Permissions service ("perm") provides authorization features for the Cloud
Foundry Platform. It answers various question forms of what particular
identities are allowed to do. It works out the answers to these questions based
on the roles assigned to users and the roles assigned to the groups they are a
member of.

Even though the service was originally created to add authorization features to
Cloud Controller, other components in the system are looking to migrate to
storing their authorization rules in Perm.

## In Popular Culture

This repository is complemented by 2 other repositories.

* [perm-release](https://github.com/cloudfoundry-incubator/perm-release)

  This is the BOSH release for deploying the `perm` service.

* [perm-rb](https://github.com/cloudfoundry-incubator/perm-rb)

  This is the Ruby library for interacting with `perm`. It is used by Cloud
  Controller to perform administration and checking of permissions.

## Usage

Not yet, please.
