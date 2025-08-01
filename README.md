# Skupper Version 2

Skupper allows you to create a Virtual Application Network (VAN) enabling secure, location independent
communication between systems including public cloud, private cloud, virtual machines (VMs),
bare metal hosts, and mainframes.

Version 1 of Skupper, [v1 branch](https://github.com/landry-some/skupper/tree/v1) is working in many production 
environments and has significantly reduced the time, effort and expense of deploying applications to a hybrid multicloud.

The main branch focuses on the development of the upcoming major release of the Skupper project based on feedback from 
users.

The plan is to produce a number of "previews" on the branch in order to get further user feedback and refine the
implementation of this major release. The v2 version is intended for evaluation purposes only and should not be used
in production environments.

# Highlights

The objective of the next Skupper major release is to better support a full declarative model so that applications
and VANs can be more easily deployed in fully automated frameworks.

The release includes:

* The introducton of Custom Resource Definitions [CRDs](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
  for a more intuitive and flexible declarative interfaces for users, with an equivalent implementation for Linux.
* Architectural improvements for the primary components (e.g. controller, cli, and non-kube executable)
* A flexible PKI implementation allowing users to easily provide their own certificates as required
* A network collector and console that is deployed separately from the site components
* Simpler integrations for centralized application network definition

# Interoperability with Version 1

Skupper v2 sites are not interoperable with v1 sites. The plan is to provide tools to assist users to
migrate their v1 installations to a v2 deployment as the release approaches.

Skupper v1 will continue to be maintained but no new significant features are planned.
