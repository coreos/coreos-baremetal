
# Getting Started

These guides walk through booting and configuring CoreOS on real or virtual hardware on your network. This introductory document reviews network booting protocols and the requirments of your network environment.

Once you've reviewed the protocols, proceed to the [boot config service](bootcfg.md) which explains a way to serve boot and cloud configurations to groups of machines. Finally, you'll be ready to use the [libvirt guide](virtual-hardware.md) or [baremetal guide](physical-hardware.md) to walk through setting up a virtual or physical network of machines.

## PXE

The Preboot eXecution Environment (PXE) defines requirements for consistent, hardware-independent network-based machine booting and configuration. Formally, PXE specifies pre-boot protocol services that client NIC firmware must provide (DHCP, TFTP, UDP/IP), specifies boot firmware requirements, and defines a client-server protocol for obtaining a network boot program (NBP) which automates OS installation and configuration.

<img src='img/pxelinux.png' class="img-center" alt="Basic PXE client server protocol flow"/>

At power-on, if a client machine's BIOS or UEFI boot firmware is set to perform network booting, the network interface card's PXE firmware broadcasts a DHCPDISCOVER packet identifying itself as a PXEClient to the network environment.

The network environment can be set up in a number of ways, which we'll discuss. In the simplest, a PXE-enabled DHCP Server responds with a DHCPOFFER with Options, which include a TFTP server IP ("next server") and the name of an NBP ("boot filename") to download (e.g. pxelinux.0). PXE firmware then downloads the NBP over TFTP and starts it. Then, the NBP loads configs, scripts, and/or images it requires to run an OS.

### Network Boot Programs

Machines can be booted and configured with CoreOS using several network boot programs and approaches. Let's review them. If you're new to network booting or unsure which to choose, iPXE is reasonable and flexible choice.

#### PXELINUX

[PXELINUX](http://www.syslinux.org/wiki/index.php/PXELINUX) is a common network boot program which tries to load a config file from the `mybootdir/pxelinux.cfg/` directory over TFTP. The file is chosen based on the client's UUID, MAC address, IP address, or a default.

    mybootdir/pxelinux.cfg/b8945908-d6a6-41a9-611d-74a6ab80b83d
    mybootdir/pxelinux.cfg/default

Here is an example PXE config file which boots a CoreOS image hosted on the TFTP server.

```
default coreos
prompt 1
timeout 15

display boot.msg

label coreos
  menu default
  kernel coreos_production_pxe.vmlinuz
  append initrd=coreos_production_pxe_image.cpio.gz cloud-config-url=http://example.com/pxe-cloud-config.yml
```

PXELINUX then downloads the specified kernel and init RAM filesystem images with TFTP.

This approach has a number of drawbacks. TFTP can be slow, managing config files can be tedious, and using different cloud configs on different machines requires separate static configs. These limitations spurred the development of various enhancements to PXE, discussed next.

In these guides, PXE is used to load the iPXE boot file so iPXE can chainload scripts and HTTP images over HTTP. Continue to the [libvirt guide](virtual-hardware.md) or the [baremetal guide](physical-hardware.md) to boot PXE clients by chainloading iPXE. Consult [CoreOS with PXE](https://coreos.com/os/docs/latest/booting-with-pxe.html) for details about CoreOS support for PXE.

#### iPXE

[iPXE](http://ipxe.org/) is an enhanced implementation of the PXE client firmware and a network boot program which uses iPXE scripts rather than config files and can download scripts and images with HTTP.

<img src='img/ipxe.png' class="img-center" alt="iPXE client server protocol flow"/>

A DHCPOFFER to iPXE client firmware specifies an HTTP boot script such as `http://example.provisioner.net/boot.ipxe`.

Here is an example iPXE script for booting the remote CoreOS stable image.

```
#!ipxe

set base-url http://stable.release.core-os.net/amd64-usr/current
kernel ${base-url}/coreos_production_pxe.vmlinuz cloud-config-url=http://provisioner.example.net/cloud-config.yml
initrd ${base-url}/coreos_production_pxe_image.cpio.gz
boot
```

A TFTP server is used only to provide the `undionly.kpxe` boot program to older PXE firmware in order to bootstrap into iPXE.

The [boot config service](bootcfg.md) can serve iPXE scripts to machines based on hardware attributes. Setup involves configuring DHCP to send iPXE clients the correct boot script endpoint.

Continue to the [libvirt guide](virtual-hardware.md) or the [baremetal guide](physical-hardware.md) to use iPXE to boot PXE/iPXE client machines. Consult [CoreOS with iPXE](https://coreos.com/os/docs/latest/booting-with-ipxe.html) for details about CoreOS support for iPXE.

#### Pixiecore

[Pixiecore](https://github.com/danderson/pixiecore) is a newer service which implements a proxyDHCP server, TFTP server, and HTTP server all-in-one and calls through to an HTTP API. The [boot config service](bootcfg.md) implements the Pixiecore API spec to provide JSON boot configs to Pixiecore based on client MAC addresses.

Continue to the [libvirt guide](virtual-hardware.md) to use Pixiecore to network boot PXE client machines.

## Network Environments

### DHCP

Many networks have DHCP services which are impractical to modify or disable. Corporate DHCP servers are governed by network admin policies and home/office networks often have routers running a DHCP service which cannot supply PXE options to PXE clients.

To address this, PXE client firmware listens for a DHCPOFFER from non-PXE DHCP server *and* a DHCPOFFER from a PXE-enabled **proxyDHCP server** which is configured to respond with just the next server and boot filename. The client firmware combines the two responses as if they had come from a single DHCP server which provided PXE Options.

<img src='img/proxydhcp.png' class="img-center" alt="DHCP and proxyDHCP responses are merged to get PXE Options"/>

The [libvirt guide](virtual-hardware.md) shows how to setup a network environment with a standalone PXE-enabled DHCP server or with a separate DHCP server and proxyDHCP server.

The [baremetal guide](physical-hardware.md) shows how to check your network environment and run either a standalone PXE-enabled DHCP server or a proxyDHCP server to compliment your existing network DHCP service.

## Configuration Service

Now that you understand network booting protocols you can explore how a [boot config service](bootcfg.md) supports PXE, iPXE, and Pixiecore network environments.
