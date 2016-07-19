
# bootcfg

`bootcfg` is an HTTP and gRPC service that renders signed [Ignition configs](https://coreos.com/ignition/docs/latest/what-is-ignition.html), [cloud-configs](https://coreos.com/os/docs/latest/cloud-config.html), network boot configs, and metadata to machines to create CoreOS clusters. `bootcfg` maintains **Group** definitions which match machines to *profiles* based on labels (e.g. UUID, MAC address, stage, region). A **Profile** is a named set of config templates (e.g. iPXE, GRUB, Ignition config, Cloud-Config). The aim is to use CoreOS Linux's early-boot capabilities to provision CoreOS machines.

Network boot endpoints provide PXE, iPXE, GRUB, and [Pixiecore](https://github.com/danderson/pixiecore/blob/master/README.api.md) support. `bootcfg` can be deployed as a binary, as an [appc](https://github.com/appc/spec) container with rkt, or as a Docker container.

<img src='img/overview.png' class="img-center" alt="Bootcfg Overview"/>

## Getting Started

Get started running `bootcfg` on your Linux machine, with rkt or Docker, to network boot virtual or physical machines into CoreOS clusters.

* [bootcfg with rkt](getting-started-rkt.md)
* [bootcfg with Docker](getting-started-docker.md)

## Flags

See [flags and variables](config.md)

## API

* [HTTP API](api.md)
* [gRPC API](https://godoc.org/github.com/mikeynap/coreos-baremetal/bootcfg/client)

## Data

A `Store` stores machine Profiles, Groups, Ignition configs, and cloud-configs. By default, `bootcfg` uses a `FileStore` to search a `-data-path` for these resources.

Prepare `/var/lib/bootcfg` with `profile`, `groups`, `ignition`, and `cloud` subdirectories. You may wish to keep these files under version control. The [examples](../examples) directory is a valid target with some pre-defined configs and templates.

     /var/lib/bootcfg
     ├── cloud
     │   ├── cloud.yaml.tmpl
     │   └── worker.sh.tmpl
     ├── ignition
     │   └── raw.ign
     │   └── etcd.yaml.tmpl
     │   └── simple.yaml.tmpl
     ├── generic
     │   └── config.yaml
     │   └── setup.cfg
     │   └── datacenter-1.tmpl
     ├── groups
     │   └── default.json
     │   └── node1.json
     │   └── us-central1-a.json
     └── profiles
         └── etcd.json
         └── worker.json

### Profiles

Profiles specify a Ignition config, Cloud-Config, and network boot config. Generic configs can be used as well.

    {
        "id": "etcd",
        "name": "CoreOS with etcd2"
        "cloud_id": "",
        "ignition_id": "etcd.yaml",
        "generic_id": "some-service.cfg",
        "boot": {
            "kernel": "/assets/coreos/899.6.0/coreos_production_pxe.vmlinuz",
            "initrd": ["/assets/coreos/899.6.0/coreos_production_pxe_image.cpio.gz"],
            "cmdline": {
                "cloud-config-url": "http://bootcfg.foo/cloud?uuid=${uuid}&mac=${net0/mac:hexhyp}",
                "coreos.autologin": "",
                "coreos.config.url": "http://bootcfg.foo/ignition?uuid=${uuid}&mac=${net0/mac:hexhyp}",
                "coreos.first_boot": "1"
            }
        }
    }

The `"boot"` settings will be used to render configs to network boot programs such as iPXE, GRUB, or Pixiecore. You may reference remote kernel and initrd assets or [local assets](#assets).

To use cloud-config, set the `cloud-config-url` kernel option to reference the `bootcfg` [Cloud-Config endpoint](api.md#cloud-config), which will render the `cloud_id` file.

To use Ignition, set the `coreos.config.url` kernel option to reference the `bootcfg` [Ignition endpoint](api.md#ignition-config), which will render the `ignition_id` file. Be sure to add the `coreos.first_boot` option as well.

#### Configs

Profiles can reference various templated configs. Ignition JSON configs can be generated from [Fuze config](https://github.com/coreos/fuze/blob/master/doc/configuration.md) template files. Cloud-Config templates files can be used to render a script or Cloud-Config. Generic template files configs can be used to render arbitrary untyped configs. Each template may contain [Go template](https://golang.org/pkg/text/template/) elements which will be executed with machine Group [metadata](#groups-and-metadata). For details and examples:

* [Ignition Config](ignition.md)
* [Cloud-Config](cloud-config.md)

## Groups and Metadata

Groups define selectors which match zero or more machines. Machine(s) matching a group will boot and provision according to the group's `Profile` and `metadata`.

Create a group definition with a `Profile` to be applied, selectors for matching machines, and any `metadata` needed to render the Ignition or Cloud config templates. For example `/var/lib/bootcfg/groups/node1.json` matches a single machine with MAC address `52:54:00:89:d8:10`.

    # /var/lib/bootcfg/groups/node1.json
    {
      "name": "node1",
      "profile": "etcd",
      "selector": {
        "mac": "52:54:00:89:d8:10"
      },
      "metadata": {
        "fleet_metadata": "role=etcd,name=node1",
        "etcd_name": "node1",
        "etcd_initial_cluster": "node1=http://172.15.0.21:2380,node2=http://172.15.0.22:2380,node3=http://172.15.0.23:2380"
      }
    }

Meanwhile, `/var/lib/bootcfg/groups/proxy.json` acts as the default machine group since it has no selectors.

    {
      "name": "etcd-proxy",
      "profile": "etcd-proxy",
      "metadata": {
        "fleet_metadata": "role=etcd-proxy",
        "etcd_initial_cluster": "node1=http://172.15.0.21:2380,node2=http://172.15.0.22:2380,node3=http://172.15.0.23:2380"
      }
    }

For example, a request to `/ignition?mac=52:54:00:89:d8:10` would render the Ignition template in the "etcd" `Profile`, with the machine group's metadata. A request to `/ignition` would match the default group (which has no selectors) and render the Ignition in the "etcd-proxy" Profile. Avoid defining multiple default groups as resolution will not be deterministic.

### Reserved Labels

Some labels are normalized or parsed specially because they have reserved semantic purpose.

* `uuid` - machine UUID
* `mac` - network interface physical address (MAC address)
* `hostname` - hostname reported by a network boot program
* `serial` - serial reported by a network boot program

Client's booted with the `/ipxe.boot` endpoint will introspect and make a request to `/ipxe` with the `uuid`, `mac`, `hostname`, and `serial` value as query arguments. Pixiecore can only detect MAC addresss and cannot substitute it into later config requests ([issue](https://github.com/mikeynap/coreos-baremetal/issues/36)).

## Assets

`bootcfg` can serve `-assets-path` static assets at `/assets`. This is helpful for reducing bandwidth usage when serving the kernel and initrd to network booted machines. The default assets-path is `/var/lib/bootcfg/assets` or you can pass `-assets-path=""` to disable asset serving.

    bootcfg.foo/assets/
    └── coreos
        └── VERSION
            ├── coreos_production_pxe.vmlinuz
            └── coreos_production_pxe_image.cpio.gz

For example, a `Profile` might refer to a local asset `/assets/coreos/VERSION/coreos_production_pxe.vmlinuz` instead of `http://stable.release.core-os.net/amd64-usr/VERSION/coreos_production_pxe.vmlinuz`.

See the [get-coreos](../scripts/README.md#get-coreos) script to quickly download, verify, and place CoreOS assets.

## Network

`bootcfg` does not implement or exec a DHCP/TFTP server. Use the [coreos/dnsmasq](../contrib/dnsmasq) image if you need a quick DHCP, proxyDHCP, TFTP, or DNS setup.

