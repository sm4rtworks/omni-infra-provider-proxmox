> ## Documentation Index for O2B Omni with ProxMox
> Fetch the complete documentation index at: https://docs.siderolabs.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Proxmox

> Creating Talos Kubernetes cluster using Proxmox.

export const release_v1_12 = 'v1.12.4';

export const VersionWarningBanner = () => {
  const latestVersion = "v1.12";
  const [latestUrl, setLatestUrl] = useState(null);
  const [currentVersion, setCurrentVersion] = useState(null);
  useEffect(() => {
    if (typeof window === "undefined") return;
    const {pathname, hash, search} = window.location;
    const match = pathname.match(/\/talos\/(v\d+\.\d+)\//);
    if (!match) return;
    const detectedVersion = match[1];
    if (detectedVersion === latestVersion) return;
    setCurrentVersion(detectedVersion);
    const newPath = pathname.replace(`/talos/${detectedVersion}/`, `/talos/${latestVersion}/`);
    setLatestUrl(`${newPath}${search}${hash}`);
  }, []);
  if (!latestUrl || !currentVersion) return null;
  return <div className="not-prose sticky top-6 z-50 my-6">
      <div className="border border-red-500/30 bg-red-500/10 px-4 py-3 rounded-xl">

        <div className="text-sm">
          ⚠️ You are viewing an older version of Talos ({currentVersion}).
          <a href={latestUrl} className="ml-2 underline text-red-400 hover:text-red-300 font-medium">
            View the latest version {latestVersion} →
          </a>
        </div>

      </div>
    </div>;
};

<VersionWarningBanner />

In this guide we will create a Kubernetes cluster using Proxmox.

## Video walkthrough

To see a live demo of this writeup, visit Youtube here:

<iframe width="560" height="315" src="https://www.youtube.com/embed/MyxigW4_QFM" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen />

## Installation

### How to get Proxmox

It is assumed that you have already installed Proxmox onto the server you wish to create Talos VMs on.
Visit the [Proxmox](https://www.proxmox.com/en/downloads) downloads page if necessary.

### Install talosctl

You can download `talosctl` on MacOS and Linux via:

```bash  theme={null}
brew install siderolabs/tap/talosctl
```

For manual installation and other platforms please see the [talosctl installation guide](../../getting-started/talosctl).

### Download ISO image

In order to install Talos in Proxmox, you will need the ISO image from [Image Factory](https://www.talos.dev/latest/talos-guides/install/boot-assets/#image-factory).

```bash  theme={null}
mkdir -p _out/
curl https://factory.talos.dev/image/376567988ad370138ad8b2698212367b8edcb69b5fd68c80be1f2ec7d603b4ba/<version>/metal-<arch>.iso -L -o _out/metal-<arch>.iso
```

For example version {release } for `linux` platform:

<CodeBlock lang="sh">
  {`
    mkdir -p _out/
    curl https://factory.talos.dev/image/376567988ad370138ad8b2698212367b8edcb69b5fd68c80be1f2ec7d603b4ba/${release_v1_12}/metal-amd64.iso -L -o _out/metal-amd64.iso
    `}
</CodeBlock>

### QEMU guest agent support (iso)

* If you need the QEMU guest agent so you can do guest VM shutdowns of your Talos VMs, then you will need a custom ISO
* To get this, navigate to [https://factory.talos.dev/](https://factory.talos.dev/)
* Scroll down and select your Talos version ( {release_v1_12} for example)
* Then tick the box for `siderolabs/qemu-guest-agent` and submit
* This will provide you with a link to the bare metal ISO
* The lines we're interested in are as follows

<CodeBlock lang="sh">
  {`
    Metal ISO

    amd64 ISO
      https://factory.talos.dev/image/ce4c980550dd2ab1b17bbf2b08801c7eb59418eafe8f279833297925d67c7515/${release_v1_12}/metal-amd64.iso
    arm64 ISO
      https://factory.talos.dev/image/ce4c980550dd2ab1b17bbf2b08801c7eb59418eafe8f279833297925d67c7515/${release_v1_12}/metal-arm64.iso

    Installer Image

    For the initial Talos install or upgrade use the following installer image:
    factory.talos.dev/installer/ce4c980550dd2ab1b17bbf2b08801c7eb59418eafe8f279833297925d67c7515:{release_v1_12}
    `}
</CodeBlock>

* Download the above ISO (this will most likely be `amd64` for you)
* Take note of the `factory.talos.dev/installer` URL as you'll need it later

## Upload ISO

From the Proxmox UI, select the "local" storage and enter the "Content" section.
Click the "Upload" button:

<img src="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-click-upload.png?fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=ff3a61b540b6f11eb29d473c40a7e25b" width="500px" data-og-width="1656" data-og-height="1244" data-path="talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-click-upload.png" data-optimize="true" data-opv="3" srcset="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-click-upload.png?w=280&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=a20baf54784dd95da59e1be62f6994f2 280w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-click-upload.png?w=560&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=0dea2def72d179f70e35d4384a81f587 560w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-click-upload.png?w=840&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=c9dc17134ceccf68264253ac2bb2eb09 840w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-click-upload.png?w=1100&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=1b6340aa3b5ce307cd4cc8bc12a6c65c 1100w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-click-upload.png?w=1650&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=25680c9d87f1603ff1f380a1e55c57fc 1650w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-click-upload.png?w=2500&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=105700f0f80a3176a563f9441ae8765a 2500w" />

Select the ISO you downloaded previously, then hit "Upload"

<img src="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=d70ac1bc7c3dbc686c228348623904d2" width="500px" data-og-width="1656" data-og-height="1244" data-path="talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png" data-optimize="true" data-opv="3" srcset="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?w=280&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=61d689ae5a4658a7be01b51053bcb10d 280w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?w=560&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=6db8cca84da1810ebf18eb9322b4fce8 560w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?w=840&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=bd78c931684210943fd8546a275480a6 840w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?w=1100&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=780c131c5482b5ce226bc393195903d2 1100w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?w=1650&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=e039926eed0a6b8b090b38c257838251 1650w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?w=2500&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=7d5b410d544be89fd3befa54cbe74c6b 2500w" />

## Create VMs

Before starting, familiarise yourself with the
[system requirements](../../getting-started/system-requirements) for Talos and assign VM
resources accordingly.

### Recommended baseline VM configuration

Use the following baseline settings for Proxmox VMs running Talos:

| Setting             | Recommended Value                               | Notes                                                                                                                           |
| ------------------- | ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| **BIOS**            | `ovmf` (UEFI)                                   | Modern firmware, Secure Boot support, better hardware compatibility                                                             |
| **Machine**         | `q35`                                           | Modern PCIe-based machine type with better device support                                                                       |
| **CPU Type**        | `host`                                          | Enables advanced instruction sets (AVX-512, etc.), best performance. Alternative: `kvm64` with feature flags for Proxmox \< 8.0 |
| **CPU Cores**       | 2+ (control plane), 4+ (workers)                | Minimum 2 cores required                                                                                                        |
| **Memory**          | 4GB+ (control plane), 8GB+ (workers)            | Minimum 2GB required                                                                                                            |
| **Disk Controller** | **VirtIO SCSI** (NOT "VirtIO SCSI Single")      | Single controller can cause bootstrap hangs (#11173)                                                                            |
| **Disk Format**     | Raw (performance) or QCOW2 (features/snapshots) | Raw preferred for performance                                                                                                   |
| **Disk Cache**      | Write Through (safe default)                    | Or None for clustered environments                                                                                              |
| **Network Model**   | `virtio`                                        | Paravirtualized driver, best performance (up to 10 Gbit)                                                                        |
| **EFI Disk**        | 4MB (for OVMF)                                  | Required for UEFI firmware, stores Secure Boot keys                                                                             |
| **Ballooning**      | Disabled                                        | Talos doesn't support memory hotplug                                                                                            |
| **RNG Device**      | VirtIO RNG (optional)                           | Better entropy for security                                                                                                     |

> **Important**: When configuring the disk controller, use **VirtIO SCSI** (not "VirtIO SCSI Single").
> Using "VirtIO SCSI Single" can cause bootstrap to hang or prevent disk discovery.
> See [issue #11173](https://github.com/siderolabs/talos/issues/11173) for details.

Create a new VM by clicking the "Create VM" button in the Proxmox UI:

<img src="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=d70ac1bc7c3dbc686c228348623904d2" width="500px" data-og-width="1656" data-og-height="1244" data-path="talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png" data-optimize="true" data-opv="3" srcset="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?w=280&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=61d689ae5a4658a7be01b51053bcb10d 280w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?w=560&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=6db8cca84da1810ebf18eb9322b4fce8 560w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?w=840&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=bd78c931684210943fd8546a275480a6 840w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?w=1100&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=780c131c5482b5ce226bc393195903d2 1100w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?w=1650&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=e039926eed0a6b8b090b38c257838251 1650w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-create-vm.png?w=2500&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=7d5b410d544be89fd3befa54cbe74c6b 2500w" />

Fill out a name for the new VM:

<img src="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-vm-name.png?fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=70436253d0216a850df3b931fdcf6a7d" width="500px" data-og-width="1656" data-og-height="1244" data-path="talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-vm-name.png" data-optimize="true" data-opv="3" srcset="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-vm-name.png?w=280&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=6be89a6512f141953f5ef03e7873387f 280w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-vm-name.png?w=560&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=033f8ede75c19054b71d77df78208aa9 560w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-vm-name.png?w=840&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=bdcfc1002cf15ba929d0eb8cff2e15b4 840w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-vm-name.png?w=1100&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=de7c3432bb5abf7d608be772c8dcc91c 1100w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-vm-name.png?w=1650&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=fc5c597cc0078aacfad972a8d8ca2578 1650w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-vm-name.png?w=2500&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=0591b20872329dd09a0a1d0b8add9c62 2500w" />

In the OS tab, select the ISO we uploaded earlier:

<img src="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-os.png?fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=de8dc5e285e886959a07bb3b45a2be27" width="500px" data-og-width="1656" data-og-height="1244" data-path="talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-os.png" data-optimize="true" data-opv="3" srcset="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-os.png?w=280&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=2fcb0ccd25998d8a77a1be8c07cbc21f 280w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-os.png?w=560&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=2c3ce113d716933d9961295d92cd3d69 560w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-os.png?w=840&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=c6b015b2eab14de1b061ed1de95f0f52 840w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-os.png?w=1100&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=90c8fd735a390fbf6afb487e785e517f 1100w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-os.png?w=1650&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=3f6a472e9ec007b298aabd87fb904965 1650w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-os.png?w=2500&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=b537273ca6b39c5f4c6a2b693bd3ec94 2500w" />

In the "System" tab:

* Set **BIOS** to `ovmf` (UEFI) for modern firmware and Secure Boot support
* Set **Machine** to `q35` for modern PCIe-based machine type
* Add **EFI Disk** (4MB) for persistent UEFI settings and Secure Boot key storage

In the "Hard Disk" tab:

* Set **Bus/Device** to `VirtIO SCSI` (NOT "VirtIO SCSI Single")
* Set **Storage** to your main storage pool
* Set **Format** to `Raw` (better performance) or `QCOW2` (features/snapshots)
* Set **Size** based on your workload requirements (adjust based on CSI and application needs)
* Set **Cache** to `Write Through` (safe default) or `None` for clustered environments
* Enable **Discard** (TRIM support) if using SSD storage
* Enable **SSD emulation** if using SSD storage

> **Important**: When configuring the disk controller, use **VirtIO SCSI** (not "VirtIO SCSI Single").
> Using "VirtIO SCSI Single" can cause bootstrap to hang or prevent disk discovery.
> See [issue #11173](https://github.com/siderolabs/talos/issues/11173) for details.

In the "CPU" section:

* Set **Cores** to 2+ for control planes, 4+ for workers
* Set **Sockets** to 1 (keep simple)
* Set **Type** to `host` (best performance, enables advanced instruction sets)
  * **Alternative for Proxmox \< 8.0**: Use `kvm64` with feature flags by adding to `/etc/pve/qemu-server/<vmid>.conf`:
    ```text  theme={null}
    args: -cpu kvm64,+cx16,+lahf_lm,+popcnt,+sse3,+ssse3,+sse4.1,+sse4.2
    ```
  * **Note**: `host` CPU type prevents live VM migration but provides best performance

<img src="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-cpu.png?fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=83c7177ea2041f1d4de7d197894fc9f1" width="500px" data-og-width="777" data-og-height="560" data-path="talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-cpu.png" data-optimize="true" data-opv="3" srcset="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-cpu.png?w=280&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=c4516c9971026da71626b3e1a14bcc78 280w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-cpu.png?w=560&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=7a2f282498d2a9fdb0ab41f8ab96b703 560w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-cpu.png?w=840&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=c281fbe84ab1d8b1f69605fec30fdd98 840w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-cpu.png?w=1100&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=b86092574a31b5e13945a75b7652d2cb 1100w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-cpu.png?w=1650&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=5f2340ba9d922537f0f85db8fbbcc5a6 1650w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-cpu.png?w=2500&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=020cd4f7342bfa60693d1c3bec3584d4 2500w" />

In the "Memory" section:

* Set **Memory** to 4GB+ for control planes, 8GB+ for workers (minimum 2GB required)
* **Disable Ballooning** (can cause issues with Talos memory detection)

<img src="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-ram.png?fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=8711e6438b6ad66d916b4f23ccf72a59" width="500px" data-og-width="1656" data-og-height="1244" data-path="talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-ram.png" data-optimize="true" data-opv="3" srcset="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-ram.png?w=280&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=ca6421b27a6e2b57f6b16cb852d1e3fc 280w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-ram.png?w=560&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=e55f5e36f3d1ff38ae9871d7ecfc5343 560w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-ram.png?w=840&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=9cba82128a73b156024fc02665d3f7c1 840w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-ram.png?w=1100&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=7e3a5bbc2d1a95680fb60859e45014d5 1100w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-ram.png?w=1650&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=41e111f6d28097d03ee138747379160a 1650w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-ram.png?w=2500&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=6b815754e0599ea28d5f6885def61156 2500w" />

In the "Network" section:

* Set **Model** to `virtio` (paravirtualized driver, best performance)
* Set **Bridge** to your network bridge (e.g., `vmbr0`)
* Verify the VM is set to come up on the bridge interface

<img src="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-nic.png?fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=8621e7e9e33d31f140a5c4de2dc75209" width="500px" data-og-width="1656" data-og-height="1244" data-path="talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-nic.png" data-optimize="true" data-opv="3" srcset="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-nic.png?w=280&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=822c0be53479b5e6e2b86bc47daa943d 280w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-nic.png?w=560&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=d24b808932fe893da8cd074106a8ad38 560w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-nic.png?w=840&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=5ccb0d15b9232a7ae0f1ecf8dc770a34 840w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-nic.png?w=1100&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=91fc5df9ce475537b47d5bc15d63c838 1100w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-nic.png?w=1650&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=092e42e924e6db485900947f660a16d9 1650w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-edit-nic.png?w=2500&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=4bce6b62f148f995a49948370defd279 2500w" />

> **Tip**: Enable a serial console (ttyS0) in Proxmox VM settings to see early boot logs and troubleshoot network connectivity issues.
> This is especially helpful when debugging DHCP timing or bridge configuration problems.
> Set **Serial port** to `ttyS0` in Proxmox and add `console=ttyS0` if you're customizing kernel args.

Finish creating the VM by clicking through the "Confirm" tab and then "Finish".

Repeat this process for a second VM to use as a worker node.
You can also repeat this for additional nodes desired.

> Note: Talos doesn't support memory hot plugging, if creating the VM programmatically don't enable memory hotplug on your
> Talos VM's.
> Doing so will cause Talos to be unable to see all available memory and have insufficient memory to complete
> installation of the cluster.

## Start control plane node

Once the VMs have been created and updated, start the VM that will be the first control plane node.
This VM will boot the ISO image specified earlier and enter "maintenance mode".

### With DHCP server

Once the machine has entered maintenance mode, there will be a console log that details the IP address that the node received.
Take note of this IP address, which will be referred to as `$CONTROL_PLANE_IP` for the rest of this guide.
If you wish to export this IP as a bash variable, simply issue a command like `export CONTROL_PLANE_IP=1.2.3.4`.

<img src="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode.png?fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=3f6f825d26c40a1d383e4de8f82c07ef" width="500px" data-og-width="1100" data-og-height="709" data-path="talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode.png" data-optimize="true" data-opv="3" srcset="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode.png?w=280&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=b7cfb4910315a52db55589ac0ac82c64 280w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode.png?w=560&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=20ca4d1b09a9566215910b0c3aa07c05 560w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode.png?w=840&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=527d0a8922698613323a759cb3bd76bf 840w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode.png?w=1100&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=f660f4a13936bf6613b8eac50d8a89ce 1100w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode.png?w=1650&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=c2da61f4f220a30a6484407fea3f467c 1650w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode.png?w=2500&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=e0a2056c596ab791006c48409edd0726 2500w" />

### Without DHCP server

To apply the machine configurations in maintenance mode, VM has to have IP on the network.
So you can set it on boot time manually.

<img src="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu.png?fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=4cf176f9711c63a359c245f1f364d85f" width="600px" data-og-width="629" data-og-height="188" data-path="talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu.png" data-optimize="true" data-opv="3" srcset="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu.png?w=280&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=02c223446d09e479709ee06773fca532 280w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu.png?w=560&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=df86a8daea0c9df0c497e7e1a59f7ed9 560w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu.png?w=840&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=3f1b7726aabd568076b791891581f12f 840w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu.png?w=1100&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=26c11416857af97e72a5eddfab568e6f 1100w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu.png?w=1650&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=669c5ed7e81d5fbdd5087ad3ae40235e 1650w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu.png?w=2500&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=a23140d67e8d6e77fe5e547d421df279 2500w" />

Press `e` on the boot time.
And set the IP parameters for the VM.
[Format is](https://www.kernel.org/doc/Documentation/filesystems/nfs/nfsroot.txt):

```bash  theme={null}
ip=<client-ip>:<srv-ip>:<gw-ip>:<netmask>:<host>:<device>:<autoconf>
```

For example \$CONTROL\_PLANE\_IP will be 192.168.0.100 and gateway 192.168.0.1

```bash  theme={null}
linux /boot/vmlinuz init_on_alloc=1 slab_nomerge pti=on panic=0 consoleblank=0 printk.devkmsg=on earlyprintk=ttyS0 console=tty0 console=ttyS0 talos.platform=metal ip=192.168.0.100::192.168.0.1:255.255.255.0::eth0:off
```

<img src="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu-ip.png?fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=909882cc272c73df855f3bc1cbbab4ba" width="630px" data-og-width="637" data-og-height="433" data-path="talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu-ip.png" data-optimize="true" data-opv="3" srcset="https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu-ip.png?w=280&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=46deafa3149aa6b467142327692f98ed 280w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu-ip.png?w=560&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=fa2bae2ade5f66a6f6c9519eff1e5ad8 560w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu-ip.png?w=840&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=9cc8ef8e54d283932e185fbd883f778d 840w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu-ip.png?w=1100&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=100cf034875f2d41902d9825ec6f8422 1100w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu-ip.png?w=1650&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=c0bb71d7a642899bd8c8559731583328 1650w, https://mintcdn.com/siderolabs-fe86397c/W5R1I1AzRuR73oKi/talos/v1.12/platform-specific-installations/virtualized-platforms/images/proxmox-maintenance-mode-grub-menu-ip.png?w=2500&fit=max&auto=format&n=W5R1I1AzRuR73oKi&q=85&s=ddcc9ba6a8afd0dea95434793799f214 2500w" />

Then press Ctrl-x or F10

## Generate machine configurations

With the IP address above, you can now generate the machine configurations to use for installing Talos and Kubernetes.
Issue the following command, updating the output directory, cluster name, and control plane IP as you see fit:

```bash  theme={null}
talosctl gen config talos-proxmox-cluster https://$CONTROL_PLANE_IP:6443 --output-dir _out
```

This will create several files in the `_out` directory: `controlplane.yaml`, `worker.yaml`, and `talosconfig`.

> Note: The Talos config by default will install to `/dev/sda`.
> Depending on your setup the virtual disk may be mounted differently Eg: `/dev/vda`.
> You can check for disks running the following command:
>
> ```bash  theme={null}
> talosctl get disks --insecure --nodes $CONTROL_PLANE_IP
> ```
>
> Update `controlplane.yaml` and `worker.yaml` config files to point to the correct disk location.

### QEMU guest agent support

For QEMU guest agent support, you can generate the config with the custom install image:

<CodeBlock lang="sh">
  {`
    talosctl gen config talos-proxmox-cluster https://$CONTROL_PLANE_IP:6443 --output-dir _out --install-image factory.talos.dev/installer/ce4c980550dd2ab1b17bbf2b08801c7eb59418eafe8f279833297925d67c7515:${release_v1_12}
    `}
</CodeBlock>

> **Important**: Enable QEMU Guest Agent in Proxmox **only if** you built the ISO with the `siderolabs/qemu-guest-agent` extension in **Image Factory**.
> If you're using a standard Talos ISO without this extension, leave QEMU Guest Agent disabled in Proxmox VM settings.
> Enabling it without the extension will only generate log spam and won't provide any functionality.
> See: [Image Factory](../../learn-more/image-factory) for building a custom ISO with extensions.

* If you did include the extension, go to your VM → **Options** and set **QEMU Guest Agent** to **Enabled**.

## Create control plane node

Using the `controlplane.yaml` generated above, you can now apply this config using talosctl.
Issue:

```bash  theme={null}
talosctl apply-config --insecure --nodes $CONTROL_PLANE_IP --file _out/controlplane.yaml
```

You should now see some action in the Proxmox console for this VM.
Talos will be installed to disk, the VM will reboot, and then Talos will configure the Kubernetes control plane on this VM.
The VM will remain in stage `Booting` until the bootstrap is completed in a later step.

> Note: This process can be repeated multiple times to create an HA control plane.

## Create worker node

Create at least a single worker node using a process similar to the control plane creation above.
Start the worker node VM and wait for it to enter "maintenance mode".
Take note of the worker node's IP address, which will be referred to as `$WORKER_IP`

Issue:

```bash  theme={null}
talosctl apply-config --insecure --nodes $WORKER_IP --file _out/worker.yaml
```

> Note: This process can be repeated multiple times to add additional workers.

## Using the cluster

Once the cluster is available, you can make use of `talosctl` and `kubectl` to interact with the cluster.
For example, to view current running containers, run `talosctl containers` for a list of containers in the `system` namespace, or `talosctl containers -k` for the `k8s.io` namespace.
To view the logs of a container, use `talosctl logs <container>` or `talosctl logs -k <container>`.

First, configure talosctl to talk to your control plane node by issuing the following, updating paths and IPs as necessary:

```bash  theme={null}
export TALOSCONFIG="_out/talosconfig"
talosctl config endpoint $CONTROL_PLANE_IP
talosctl config node $CONTROL_PLANE_IP
```

### Bootstrap etcd

```bash  theme={null}
talosctl bootstrap
```

### Retrieve the `kubeconfig`

At this point we can retrieve the admin `kubeconfig` by running:

```bash  theme={null}
talosctl kubeconfig .
```

## Troubleshooting

### Cluster creation issues

If `talosctl cluster create` fails with disk controller errors:

* **"virtio-scsi-single disk controller is not supported"**: This disk controller type causes Talos bootstrap to hang. Use `virtio` or `scsi` instead:
  ```bash  theme={null}
  # Wrong - will be rejected
  talosctl cluster create --disks virtio-scsi-single:10GiB

  # Correct - use virtio or scsi
  talosctl cluster create --disks virtio:10GiB
  talosctl cluster create --disks scsi:10GiB
  ```

### Network connectivity issues

If nodes fail to obtain IP addresses or show "network is unreachable" errors:

1. **Verify bridge interface**: Ensure the bridge interface (e.g., `vmbr0`) exists and is UP before starting VMs
   ```bash  theme={null}
   ip link show vmbr0
   ```

2. **Check DHCP server**: Ensure DHCP server is running and reachable from the bridge network

3. **Firewall rules**: If Proxmox VM firewall is enabled, allow DHCP traffic (UDP ports 67/68).
   If you enforce further filtering, ensure control-plane/API connectivity per your environment's policy (see Talos networking docs).

4. **VLAN configuration**: Ensure VLAN tags match between bridge configuration, VM network settings, and switch configuration

5. **Serial console**: Enable serial console to view early boot logs and network initialization messages

### Disk Controller Issues

* **Configuration rejected**: If you see "virtio-scsi-single disk controller is not supported", use `--disks virtio:10GiB` instead of `--disks virtio-scsi-single:10GiB`
* **Bootstrap hangs**: If bootstrap hangs or disks aren't discovered, verify you're using **VirtIO SCSI** (not "VirtIO SCSI Single")
* **Disk not found**: Check disk path using `talosctl get disks --insecure --nodes $CONTROL_PLANE_IP` and update `install.disk` in machine config if needed (e.g., `install.disk: /dev/vda`)

### Secure boot

For Secure Boot setup, see the [Secure Boot documentation](../bare-metal-platforms/secureboot).

## Cleaning up

To cleanup, simply stop and delete the virtual machines from the Proxmox UI.
