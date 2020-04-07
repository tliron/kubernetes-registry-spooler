Kubernetes Registry Spooler
===========================

A sidecar to update an OCI/Docker container image registry using files instead of RESTful HTTP
calls.

What? Why?
----------

You're here because you need to run your own container image registry, for whatever reason:
development, testing, security, etc. The "right" way to do it is to run the registry outside of your
Kubernetes clusters, either as a standalone service or in its own special cluster. But, often it is
convenient or useful to run the registry within the same cluster that it will be used. This is
common when working on Kubernetes-based development on a single machine, e.g. via Minikube, in which
case all you have is a single Kubernetes cluster, so why not push your container images directly to
it? But your reason might be more purposeful, e.g. using a local registry as a cache, as an
inventory for in-cluster deployments, etc.

Unfortunately, it can be difficult to access such a private registry from outside the cluster, even
if all you need is to push images to it. You would need to enable some kind of ingress for HTTP,
which could entail
full-blown [`Ingress`](https://kubernetes.io/docs/concepts/services-networking/ingress/) support,
a [`LoadBalancer`-type `Service`](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer),
a [`kubectl port-forward`](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/),
or even a [`NodePort`](https://kubernetes.io/docs/concepts/services-networking/service/#nodeport)
(a.k.a. "giving up"), all of which have complex requirements and drawbacks.

However, we can make use of a feature that will *always* be available as long as you can access
the cluster's API server: the ability to stream data to a command over its RESTful interface (via
SPDY). It is available via client libraries (e.g.
[in the Go client](https://pkg.go.dev/k8s.io/client-go/tools/remotecommand))
or via the CLI as
[`kubectl exec`](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#exec),
e.g. the `cp` shorcut:

    kubectl cp myfile mypodname:/mydirectory/

You might see where this is going: if we can copy files to the registry's pod, then we can add a
sidecar that would pick up these files and push them to the registry for us. In other words: a
spooler.

Installation
------------

Let's deploy our registry with the spooler sidecar. We include a
[sample spec](assets/registry-with-spooler.yaml) that you can use as is or modify:

    kubectl namespace create mynamespace
    kubectl apply --filename=assets/registry-with-spooler.yaml --namespace=mynamespace

As you can see, it's a very straightforward spec and you could adapt the technique of adding a
sidecar to more elaborate registry servers, such as [Quay](https://github.com/quay/quay) or
[Harbor](https://github.com/goharbor/harbor).

Also note that this spec uses our
[pre-built container image on Docker Hub](https://hub.docker.com/r/tliron/kubernetes-registry-spooler).
See our [scripts](scripts/) to learn how to build it yourself.  

Once it's up and running let's get our registry's (first) pod name:

    POD=$(kubectl get pods --selector=app.kubernetes.io/name=registry --field-selector=status.phase=Running --namespace=mynamespace --output=jsonpath={.items[0].metadata.name})

To see the spooler logs:

    kubectl logs $POD --container=spooler --namespace=mynamespace

Pushing to the Registry
-----------------------

Now we can copy files to the spooler:

    kubectl cp myimage.tar $POD:/spool --container=spooler --namespace=mynamespace

The filename, stripped of extensions, will become the container image name. E.g. `myimage.tar` will
be pushed to the repository as `myimage`. Once it is pushed the file will be deleted from the spool
directory. Check the logs (see above) to make sure everything worked as expected.

Note that though you could potentially push any kind of file, most registry implementations would
only be able to meaningfully store [tar files](https://www.gnu.org/software/tar/). That said, the
tar does *not* have to contain a container image, and indeed any content can be pushed as long as
it's tarred.  

Deleting from the Registry
--------------------------

Can we create anti-file on the spooler? Kinda! Just add "!" to the end of the filename. The content
of the file doesn't matter (and neither does the extension), so a simple `touch` command should be
enough:

    kubectl exec $POD --container=spooler --namespace=mynamespace -- touch /spool/myimage!

Pulling from the Registry
-------------------------

We provide a `registry-pull` tool for pulling files, as well as a directory to put them in. Of
course it only makes sense to execute it in the container. Pulling could thus be done in three
steps:

    kubectl exec $POD --container=spooler --namespace=mynamespace -- registry-pull myimage /pull/myimage.tar
    kubectl cp $POD:/pull/myimage.tar myimage.tar --container=spooler --namespace=mynamespace
    kubectl exec $POD --container=spooler --namespace=mynamespace -- rm /pull/myimage.tar
