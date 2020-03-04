# Secret Manager Operator (Beta)

The Secret Manager Operator provides a GitOps-friendly way to use Google Cloud's [Secrets Management](https://cloud.google.com/solutions/secrets-management) tools. It works using the `ManagedSecret` CRD, which creates a Kubernetes secret and manages its lifecycle. A ManagedSecret specifies a project, secret, and version. This is then retrieved and placed in a similarly-named Kubernetes secret. When the ManagedSecret's version is updated, the Secret it manages will be updated as well. Finally, if the ManagedSecret opts-in to this behaviour by including the finalizer `managedsecrets.endclothing.com`, the Secret will be destroyed along with it when it is deleted. 

# Installation

A basic installation is provided via Kustomize:
```
bases:
- github.com/endclothing/secret-manager-operator/manifests?ref=v0.0.1

namespace: default
```

This will create a deployment, service account, and cluster-wide RBAC for the Secret Manager Operator. On top of this, you will to provide the pod with credentials allowing the `secretmanager.versions.access` IAM permission, either by setting the `GOOGLE_APPLICATION_CREDENTIALS` environment key and providing a suitable private key file, or via [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity).

# Basic Usage
1. Create a secret in your environment like so:

```
export PROJECT="<your google project>"
export SECRET_ID="my-awful-secret"
gcloud beta secrets create "${SECRET_ID}" \
    --replication-policy="automatic" \
    --project="${PROJECT}"
```

2. Create a new version of your secret with the desired contents
```
echo -n "I ate all the cookies" > ./secret
export PROJECT="<your google project>"
export SECRET_ID="my-awful-secret"
gcloud beta secrets versions add "${SECRET_ID}" --data-file="./secret" --project="${PROJECT}"
```

3. Check your secret was recorded successfully:

```
export PROJECT="<your google project>"
export SECRET_ID="my-awful-secret"
gcloud beta secrets list --project="${PROJECT}" | grep "${SECRET_ID}"
> my-awful-secret  2020-02-17T14:27:17  automatic           -
```

4. Create a CRD to tell the Secret Manager operator to synchronise your secret with the cluster:

```
{
   "apiVersion": "endclothing.com/v1",
   "kind": "ManagedSecret",
   "metadata": {
      "name": "my-awful-secret",
   },
   "spec": {
      "generation": 1,
      "project": "<your google project>",
      "secret": "my-awful-secret"
   }
}
```

5. Optionally, check the cluster to verify your secret has been sychronised:

```
kubectl get secret -o yaml my-awful-secret | yq r - 'data.contents' | base64 -D
> I ate all the cookies
```

# Adoption Considerations
There are a couple of caveats of which you should be aware when adopting the Secret Manager Operator:
1. The operator will destructively modify existing secrets. e.g. if you have an existing secret named `my-awful-secret` with arbitrary keys and you create a ManagedSecret with the same name,
the existing contents *will be deleted*. This may cause an outage or loss of non-redundant data. For this reason, you should ensure that you do not create ManagedSecrets whose names overlap existing secrets while adopting the Secret Manager Operator.

2. You should consider carefully how the Secret Manager Operator interacts with your security model. In particular:  
    1. While the code is simple and is used in production at END., there is no public audit available
    2. The operator creates Kubernetes secrets, which may not be encrypted at rest depending on how your Kubernetes cluster is configured
    3. The operator assumes cluster-wide permissions to access secrets, which may not be appropriate in some multitenant scenarios

3. The operator does not handle renaming gracefully. If you create a ManagedSecret and later change the name of the secret it references, the operator *will* create a new secret for the new name, but it *will not* clean up the old one. You should bear this in mind if you do not have ad-hoc access to your Kubernetes cluster and you care about leaving old secrets in the environment. This can be effectively worked around by:  
  1. Creating a new ManagedSecret with the desired name 
  2. Updating your services to consume this secret
  3. Opting-in to finalization on the old ManagedSecret
  4. Deleting the old ManagedSecret

# Project Status
This project is used in production and END. Clothing and is feature-complete, however it is relatively young. It may not be bug-free and it has not been tested on a broadly-configured set of Kubernetes clusters. The code should work in most sensible cases and is likely fixable in others, but you should not rely on it working out-of-the-box the way you might with a mature tool.

# Comparison to Other Secret Management Solutions
To try and help you make an informed choice about whether to adopt the Secret Manager Operator, we include a brief list of alternative solutions and our evaluation of whether they suited us. Obviously, given that we wrote the Secret Manager Operator, our final answer was no in all cases, but our use-case may very well be different to yours and these projects are generally high-quality and worthy of your serious consideration.

## Kubernetes Secrets
The most obvious place to start managing secrets in Kubernetes is Kubernetes itself. Kubernetes secrets work well and, when configured appropriately, are a secure and reliable mechanism for storing secrets. However, for most real-world use cases, they are not a complete solution. This is because their life cycle is tied to a Kubernetes cluster. They do not offer an affordance for consuming secrets outside of a cluster, recording them beyond the life of the cluster, or ensuring they are synchronised between clusters. In practice, this means that most organisations will need a more comprehensive approach to secret management than they can offer.

## [hashicorp/vault](https://github.com/hashicorp/vault)
Vault is probably the most recognisable name in this sphere at the moment and offers a very powerful suite of features at the cost of being a complex piece of software with demanding operational requirements and a relatively limited free offering (HA or multi-cluster deployments of Vault require a commercial license, which costs at least mid five-figures). However, if you are well-resourced and can make use of its advanced features, it's pretty much the only game in town. In particular, Vault offers a number of features that are not widely available elsewhere:
1. Dynamic short-lived secrets
2. The broadest support of any solution for pluggable backends
3. A relatively full-featured GUI
4. Broad support for pluggable authentication

If you have a clear use-case for these features, Vault should probably be your first choice. If you don't, something simpler can probably solve your problem.

## [mozilla/sops](https://github.com/mozilla/sops) and [StackExchange/blackbox](https://github.com/StackExchange/blackbox)
Both sops and blackbox are powerful and flexible solutions for working with secrets stored in a VCS. Both have a solid track record of development by well-known and trustworthy companies, and both are mature. However, neither specifically targets Kubernetes as a platform. This means that using them involves doing some plumbing in your CD system to render secret values into manifests and then apply them. If this is no trouble for you, then either is a strong contender. However, if you use a GitOps tool like [argoproj/argo-cd](https://github.com/argoproj/argo-cd) or [fluxcd/flux](https://github.com/fluxcd/flux) directly, then this may not be convenient for you and you will want to consider a Kubernetes-native tool. Such as...

## [bitnami-labs/sealed-secrets](https://github.com/bitnami-labs/sealed-secrets), [godaddy/kubernetes-external-secrets](https://github.com/godaddy/kubernetes-external-secrets), [Solute/kamus](https://github.com/Soluto/kamus), [futuresimple/helm-secrets](https://github.com/futuresimple/helm-secrets), [GoogleCloudPlatform/berglas](https://github.com/GoogleCloudPlatform/berglas), and numerous others
There are too many Kubernetes-native offerings to list or evaluate them all here, so this is just a quick overview of the solutions we considered using and ultimately passed over in favour of writing our own along with our reasoning:

### [bitnami-labs/sealed-secrets](https://github.com/bitnami-labs/sealed-secrets)
* Requires access to Kubernetes API
* Unclear story around master key rotation and multiple clusters
* Unclear ongoing maintenance and development situation 

### [godaddy/kubernetes-external-secrets](https://github.com/godaddy/kubernetes-external-secrets)
* Doesn't support GCP KMS
* Written in JavaScript

### [Soluto/kamus](https://github.com/Soluto/kamus)
* Complicated deployment structure (runs three separate services)
* Mysterious health check failures which we struggled to debug
* Written in .NET

### [futuresimple/helm-secrets](https://github.com/futuresimple/helm-secrets)
* We don't use Helm

### [GoogleCloudPlatform/berglas](https://github.com/GoogleCloudPlatform/berglas)
* Berglas is only compatible with Kubernetes pods that specify a command expliclity, and we wanted to avoid having to perform YAML surgery on third-party tools we used
* Berglas relies on a MutatingWebhook, and webhooks in Kubernetes tend to (not) compose in ways that lead to [extremely confusing bugs](https://github.com/istio/istio/issues/17318). Thus we would prefer a solution that does not rely on one if possible
