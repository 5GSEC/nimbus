# Nimbus API

This document provides guidance on extending and maintaining the [Nimbus API](../../api)

## Concepts

* https://kubernetes.io/docs/reference/using-api/api-concepts/
* https://kubernetes.io/docs/reference/using-api/
* https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/
* https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md

## API Groups

All Nimbus resources are currently defined in the `intent.security.nimbus.com` API group.

## API Versions

This `intent.security.nimbus.com` has the following versions:

* v1alpha1

## Adding a new attribute

New attributes can be added to existing resources without impacting compatibility. They do not require a new version.

## Deleting an attribute

Attributes cannot be deleted in a version. They should be marked for deprecation and removed after 3 releases.

## Modifying an attribute

Attributes cannot be modified in a version. The existing attribute should be marked for deprecation and a new attribute
should be added following version compatibility guidelines.
