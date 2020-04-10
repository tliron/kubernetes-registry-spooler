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
if all you need is to push images to it. You would need to enable some kind of way in for HTTP,
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
sidecar that would pick up these files and locally push them to the registry for us. In other words:
a spooler.

Installation
------------

Let's deploy our registry with the spooler sidecar. We include a
[sample spec](assets/registry-with-spooler.yaml) that you can use as is or modify:

    kubectl create namespace mynamespace
    kubectl apply --filename=assets/registry-with-spooler.yaml --namespace=mynamespace

As you can see, it's a very straightforward spec and you could adapt the technique of adding a
sidecar to more elaborate registry servers, such as [Quay](https://github.com/quay/quay) or
[Harbor](https://github.com/goharbor/harbor).

Also note that this spec uses our
[pre-built container image on Docker Hub](https://hub.docker.com/r/tliron/kubernetes-registry-spooler).
See our [scripts](scripts/) to learn how to build it yourself.  

Once it's up and running let's get our registry's (first) pod name:

    POD=$(kubectl get pods --selector=app.kubernetes.io/name=registry --namespace=mynamespace --output=jsonpath={.items[0].metadata.name})

To see the spooler logs:

    kubectl logs $POD --container=spooler --namespace=mynamespace

Pushing to the Registry
-----------------------

Now we can copy files to the spooler:

    echo 'hello world' > /tmp/hello.txt
    kubectl cp /tmp/hello.txt $POD:/spool/hello.txt~ --container=spooler --namespace=mynamespace
    kubectl exec $POD --container=spooler --namespace=mynamespace -- mv /spool/hello.txt~ /spool/hello.txt

Note that we are copying and then renaming the file. The reason is that we don't want the spooler
to push the file before we are done writing to it. The spooler ignores files ending with "~".

The filename, stripped of extensions, will become the container image name. E.g. `hello.txt` will
be pushed to the repository as `hello`. Once it is pushed the file will be deleted from the spool
directory. Check the logs (see above) to make sure everything worked as expected.

Note that you could potentially push *any* kind of file to the registry, as in this example of a
text file. In this case the registry would assume that you are sending a raw (uncompressed) image
layer.

However, a *real* image would be a tarball with a `manifest.json`, a `sha256:` file, and the layers.
The spooler treats files with the `.tar` extension as such.

How would you go about creating such tarballs? You can use a tool like [podman](https://podman.io/).
For example, let's save a tarball from Docker Hub:

    podman pull registry.hub.docker.com/library/registry
    podman tag registry.hub.docker.com/library/registry localhost:5000/myregistry
    podman save localhost:5000/myregistry --output /tmp/myregistry.tar

Note that we had to add a tag to the image so that it would match the internal push that the spooler
will do. And now let's push it, exactly the same way as before:

    kubectl cp /tmp/myregistry.tar $POD:/spool/myregistry.tar~ --container=spooler --namespace=mynamespace
    kubectl exec $POD --container=spooler --namespace=mynamespace -- mv /spool/myregistry.tar~ /spool/myregistry.tar

Deleting from the Registry
--------------------------

Can we create anti-file on the spooler? Kinda! Just add "!" to the end of the filename. The content
of the file doesn't matter (and neither does the extension), so a simple `touch` command should be
enough:

    kubectl exec $POD --container=spooler --namespace=mynamespace -- touch /spool/hello!

Pulling from the Registry
-------------------------

The sidecar has a `registry-pull` tool for pulling the tarball to stdout:

    kubectl exec $POD --container=spooler --namespace=mynamespace -- registry-pull hello > /tmp/hello.tar

Note that the pulled file would *always* be a tarball, so it's a good idea to always use the `.tar`
extension. You could untar it like so:

    tar --extract --verbose --file=/tmp/hello.tar

That should extract a `manifest.json`, a `sha256:` file, as well as a single compressed layer with a
`.tar.gz` extension. We could extract our original text file content like so:

    cat c6b8929c27b26f5c6f322583f20183b804afc613b9af545502d8bce40d025fdf.tar.gz | gunzip

(Replace the filename with what was extracted from `hello.tar`)

Note that in our example the `.tar` extension before the `.gz` is misleading, because our layer is
just a raw text file, not an actual tarball.

Client API
----------

All the above could be fine for scripting, but also provided here is a Go API that does it all
programmatically. See the [client/](client/) directory. You can import it directly into your
program:

```go
import (
    spooler "github.com/tliron/kubernetes-registry-spooler/client"
)
```
