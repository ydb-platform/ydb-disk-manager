## ydb-disk-manager 

This is the helper process that solves a Kubernetes-specific problem. 
It is intended to be used as a DaemonSet in Kubernetes installations of YDB to make sure YDB has access to disks instead of EPERM.

The codebase is based upon `smarter-device-manager` with vital modifications.

#### ListAndWatch:

The only device advertised to kubelet is `ydb-disk-manager/hostdev`, which is a metadevice that is considered to be always present
on every node (since `/dev` path is always present on every node)

#### Allocate:

`Allocate` response returns a series of disks instead the metadevice `ydb-disk-manager/hostdev` that was allocated. Here, we abuse kubelet
behaviour - kubelet will silently swallow every device provided, will not check if it is the same device that it requested, and will propagate it
down to the container runtime.
