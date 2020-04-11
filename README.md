Kubernetes Registry Spooler
===========================

A sidecar to update an OCI/Docker container image registry using files instead of RESTful HTTP
calls.

The Problem
-----------

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

The Solution
------------

We can make use of a feature that will *always* be available as long as you can access the cluster's
API server: the ability to stream data to a command over its RESTful interface (via SPDY). It's
supported by client libraries (e.g. [in the Go client](https://pkg.go.dev/k8s.io/client-go/tools/remotecommand))
and even in the CLI as
[`kubectl exec`](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#exec)
and the related `kubectl cp` shortcut:

    kubectl cp myfile mypodname:/mydirectory/

You might see where this is going: if we can copy files to the registry's pod, then we can add a
sidecar that would pick up these files and push them to the in-pod registry for us. In other words:
a spooler.

(Small quibble: actually `kubectl cp` specifically requires the `tar` tool to be available in the
remote container, so it won't work absolutely everywhere. This is
[a known issue](https://github.com/kubernetes/kubernetes/issues/58512). However, the Go client does
not have this limitation and you can choose your remote command.) 

Installation
------------

Let's deploy our registry with the spooler sidecar:

    kubectl create namespace mynamespace
    scripts/install mynamespace

We are using our [sample spec](assets/registry-with-spooler.yaml) that you can use as is or modify.
As you can see, it's a very straightforward spec and you could adapt the technique of adding a
sidecar to more elaborate registry servers, such as [Quay](https://github.com/quay/quay) or
[Harbor](https://github.com/goharbor/harbor). To install:

Also note that this spec uses our
[pre-built container image on Docker Hub](https://hub.docker.com/r/tliron/kubernetes-registry-spooler).
See our [development](development/) directory to learn how to build it yourself.  

We can now see the spooler logs:

    scripts/logs mynamespace

Note that for all our scripts the final namespace argument is optional. If not specified, it will
use the currently configured namespace. You can configure that:

    scripts/namespace mynamespace

Pushing to the Registry
-----------------------

    echo 'hello world' > /tmp/hello.txt
    scripts/push /tmp/hello.txt hello mynamespace

If you look into [`scripts/push`](scripts/push) you'll see that we are copying and then renaming the
file. The reason is that we don't want the spooler to push the file before we are done writing to
it. The spooler ignores files ending with "~".

The filename in `/spool/`, stripped of extensions, will become the container image name. E.g.
`hello.txt` will be pushed to the repository as `hello`. Note that backslashes will be converted to
slashes for the image name (we do this because slashes cannot be used in filenames). 

Once it is pushed the file will be deleted from the spool directory. Check the logs (see above) to
make sure everything worked as expected.

Note that you could potentially push *any* kind of file to the registry, as in this example of a
text file. In this case the registry would assume that you are sending a raw (uncompressed) image
layer.

However, a *real* image would be a tarball with a `manifest.json`, a `sha256:` file, and the layers.
The spooler treats files with the `.tar` extension as such.

How would you go about creating such tarballs? You can use a tool like [podman](https://podman.io/).
For example, let's save a tarball from Docker Hub:

    podman pull registry.hub.docker.com/library/hello-world
    podman tag registry.hub.docker.com/library/hello-world localhost:5000/catalog/myimage
    podman save localhost:5000/catalog/myimage --output /tmp/myimage.tar

Note that we had to add a tag to the image so that it would match the internal push that the spooler
will do. And now let's push it, exactly the same way as before:

    scripts/push /tmp/myimage.tar catalog/myimage mynamespace

Deleting from the Registry
--------------------------

    scripts/delete hello mynamespace

How does this work? Can we create anti-file on the spooler? Kinda! We just add "!" to the end of the
filename, which the spooler interprets to mean deletion. The content of the file doesn't matter (and
neither does the extension), so a simple `touch` is enough.

Pulling from the Registry
-------------------------

    scripts/pull hello mynamespace > hello.tar

This works by using a `registry` tool that we've included in the sidecar.

Note that the pulled file would *always* be a tarball, so it's a good idea to always use the `.tar`
extension, as we did here. You could pull and untar in one line, like so:

    scripts/pull hello mynamespace | tar --extract --verbose

That should extract a `manifest.json`, a `sha256:` file, as well as a single compressed layer with a
`.tar.gz` extension. Note that in our example the `.tar` extension before the `.gz` is misleading,
because our layer is just a raw text file, not an actual tarball.

We can untar that one layer and unzip it in one line:

    scripts/pull hello mynamespace | tar --extract --to-stdout *.gz | gunzip

Listing the Contents of the Registry
------------------------------------

    scripts/list mynamespace

This again uses the `registry` tool in the sidecar.

Client API
----------

All these scripts could be fine for scripting, but also provided here is a Go API that does it all
programmatically. See the [client/](client/) directory. You can import it directly into your
program:

```go
import (
    spooler "github.com/tliron/kubernetes-registry-spooler/client"
)
```
