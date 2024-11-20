The Google compute Packer plugin lets you create custom images for use within Google Compute Engine (GCE).

### Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    googlecompute = {
      source  = "github.com/hashicorp/googlecompute"
      version = "~> 1"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/hashicorp/googlecompute
```

### Components

#### Builders

- [googlecompute](/packer/integrations/hashicorp/googlecompute/latest/components/builder/googlecompute) - The
  googlecompute builder creates images from existing ones, by launching an instance, provisioning it, then exporting
  it as a reusable image.

#### Post-Processors

- [googlecompute-import](/packer/integrations/hashicorp/googlecompute/latest/components/post-processor/googlecompute-import) -
  The googlecompute-import post-processor imports an existing raw disk image, and imports it as a GCE image that can be
  used for launching instances from.

- [googlecompute-export](/packer/integrations/hashicorp/googlecompute/latest/components/post-processor/googlecompute-export) -
  The googlecompute-export post-processor exports the image built by the googlecompute builder as a .tar.gz archive into Google
  Cloud Storage (GCS).

### Authentication

Authenticating with Google Cloud services requires either a User Application Default Credentials,
a JSON Service Account Key or an Access Token.  These are **not** required if you are
running the `googlecompute` Packer builder on Google Cloud with a
properly-configured [Google Service
Account](https://cloud.google.com/compute/docs/authentication).

The following options are available for the `googlecompute` builder, the `googlecompute-export`, and
the `googlecompute-import`:

@include 'lib/common/Authentication-not-required.mdx'

#### Running locally on your workstation.

If you run the `googlecompute` Packer builder locally on your workstation, you will
need to install the Google Cloud SDK and authenticate using [User Application Default
Credentials](https://cloud.google.com/sdk/gcloud/reference/auth/application-default).
You don't need to specify an _account file_ if you are using this method. Your user
must have at least `Compute Instance Admin (v1)` & `Service Account User` roles
to use Packer succesfully.

#### Running on Google Cloud

If you run the `googlecompute` Packer builder on GCE or GKE, you can
configure that instance or cluster to use a [Google Service
Account](https://cloud.google.com/compute/docs/authentication). This will allow
Packer to authenticate to Google Cloud without having to bake in a separate
credential/authentication file.

It is recommended that you create a custom service account for Packer and assign it
`Compute Instance Admin (v1)` & `Service Account User` roles.

For `gcloud`, you can run the following commands:

```shell-session
$ gcloud iam service-accounts create packer \
  --project YOUR_GCP_PROJECT \
  --description="Packer Service Account" \
  --display-name="Packer Service Account"

$ gcloud projects add-iam-policy-binding YOUR_GCP_PROJECT \
    --member=serviceAccount:packer@YOUR_GCP_PROJECT.iam.gserviceaccount.com \
    --role=roles/compute.instanceAdmin.v1

$ gcloud projects add-iam-policy-binding YOUR_GCP_PROJECT \
    --member=serviceAccount:packer@YOUR_GCP_PROJECT.iam.gserviceaccount.com \
    --role=roles/iam.serviceAccountUser

$ gcloud projects add-iam-policy-binding YOUR_GCP_PROJECT \
    --member=serviceAccount:packer@YOUR_GCP_PROJECT.iam.gserviceaccount.com \
    --role=roles/iap.tunnelResourceAccessor

$ gcloud compute instances create INSTANCE-NAME \
  --project YOUR_GCP_PROJECT \
  --image-family ubuntu-2004-lts \
  --image-project ubuntu-os-cloud \
  --network YOUR_GCP_NETWORK \
  --zone YOUR_GCP_ZONE \
  --service-account=packer@YOUR_GCP_PROJECT.iam.gserviceaccount.com \
  --scopes="https://www.googleapis.com/auth/cloud-platform"
```

**The service account will be used automatically by Packer as long as there is
no _account file_ specified in the Packer configuration file.**

#### Running outside of Google Cloud

The [Google Cloud Console](https://console.cloud.google.com) allows
you to create and download a credential file that will let you use the
`googlecompute` Packer builder anywhere. To make the process more
straightforwarded, it is documented here.

1.  Log into the [Google Cloud
    Console](https://console.cloud.google.com/iam-admin/serviceaccounts) and select a project.

2.  Click Select a project, choose your project, and click Open.

3.  Click Create Service Account.

4.  Enter a service account name (friendly display name), an optional description, select the `Compute Engine Instance Admin (v1)` and `Service Account User` roles, and then click Save.

5.  Generate a JSON Key and save it in a secure location.

6.  Set the Environment Variable `GOOGLE_APPLICATION_CREDENTIALS` to point to the path of the service account key.

#### Precedence of Authentication Methods

Packer looks for credentials in the following places, preferring the first
location found:

1.  An `access_token` option in your packer file.

2.  An `account_file` option in your packer file.

3.  A JSON file (Service Account) whose path is specified by the
    `GOOGLE_APPLICATION_CREDENTIALS` environment variable.

4.  A JSON file in a location known to the `gcloud` command-line tool.
    (`gcloud auth application-default login` creates it)

    On Windows, this is:

        %APPDATA%/gcloud/application_default_credentials.json

    On other systems:

        $HOME/.config/gcloud/application_default_credentials.json

5.  On Google Compute Engine and Google App Engine Managed VMs, it fetches
    credentials from the metadata server. (Needs a correct VM authentication
    scope configuration, see above.)

### Examples

#### Basic Example

Below is a fully functioning example. It doesn't do anything useful since no
provisioners or startup-script metadata are defined, but it will effectively
repackage an existing GCE image.

**JSON**

```json
{
  "builders": [
    {
      "type": "googlecompute",
      "project_id": "my project",
      "source_image": "debian-9-stretch-v20200805",
      "ssh_username": "packer",
      "zone": "us-central1-a"
    }
  ]
}
```

**HCL2**

```hcl
source "googlecompute" "basic-example" {
  project_id = "my project"
  source_image = "debian-9-stretch-v20200805"
  ssh_username = "packer"
  zone = "us-central1-a"
}

build {
  sources = ["sources.googlecompute.basic-example"]
}
```


#### Windows Example

Before you can provision using the winrm communicator, you need to allow
traffic through google's firewall on the winrm port (tcp:5986). You can do so
using the gcloud command.

    gcloud compute firewall-rules create allow-winrm --allow tcp:5986

Or alternatively by navigating to [https://console.cloud.google.com/networking/firewalls/list](https://console.cloud.google.com/networking/firewalls/list).

Once this is set up, the following is a complete working packer config after
setting a valid `project_id`:

**JSON**

```json
{
  "builders": [
    {
      "type": "googlecompute",
      "project_id": "my project",
      "source_image": "windows-server-2019-dc-v20200813",
      "disk_size": "50",
      "machine_type": "n1-standard-2",
      "communicator": "winrm",
      "winrm_username": "packer_user",
      "winrm_insecure": true,
      "winrm_use_ssl": true,
      "metadata": {
        "sysprep-specialize-script-cmd": "winrm quickconfig -quiet & net user /add packer_user & net localgroup administrators packer_user /add & winrm set winrm/config/service/auth @{Basic=\"true\"}"
      },
      "zone": "us-central1-a"
    }
  ]
}
```

**HCL2**

```hcl
source "googlecompute" "windows-example" {
  project_id = "MY_PROJECT"
  source_image = "windows-server-2019-dc-v20200813"
  zone = "us-central1-a"
  disk_size = 50
  machine_type = "n1-standard-2"
  communicator = "winrm"
  winrm_username = "packer_user"
  winrm_insecure = true
  winrm_use_ssl = true
  metadata = {
    sysprep-specialize-script-cmd = "winrm quickconfig -quiet & net user /add packer_user & net localgroup administrators packer_user /add & winrm set winrm/config/service/auth @{Basic=\"true\"}"
  }
}

build {
  sources = ["sources.googlecompute.windows-example"]
}
```

-> **Warning:** Please note that if you're setting up WinRM for provisioning, you'll probably want to turn it off or restrict its permissions as part of a shutdown script at the end of Packer's provisioning process. For more details on the why/how, check out this useful blog post and the associated code:
https://missionimpossiblecode.io/post/winrm-for-provisioning-close-the-door-on-the-way-out-eh/

This build can take up to 15 min.

#### Windows over WinSSH Example

The following uses Windows SSH as backend communicator
[https://docs.microsoft.com/en-us/windows-server/administration/openssh/openssh_install_firstuse](https://docs.microsoft.com/en-us/windows-server/administration/openssh/openssh_install_firstuse)

```hcl
source "googlecompute" "windows-ssh-example" {
  project_id = "MY_PROJECT"
  source_image = "windows-server-2019-dc-v20200813"
  zone = "us-east4-a"
  disk_size = 50
  machine_type = "n1-standard-2"
  communicator = "ssh"
  ssh_username = var.packer_username
  ssh_password = var.packer_user_password
  ssh_timeout = "1h"
  metadata = {
    sysprep-specialize-script-cmd = "net user ${var.packer_username} \"${var.packer_user_password}\" /add /y & wmic UserAccount where Name=\"${var.packer_username}\" set PasswordExpires=False & net localgroup administrators ${var.packer_username} /add & powershell Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0 & powershell Start-Service sshd & powershell Set-Service -Name sshd -StartupType 'Automatic' & powershell New-NetFirewallRule -Name 'OpenSSH-Server-In-TCP' -DisplayName 'OpenSSH Server (sshd)' -Enabled True -Direction Inbound -Protocol TCP -Action Allow -LocalPort 22 & powershell.exe -NoProfile -ExecutionPolicy Bypass -Command \"Set-ExecutionPolicy -ExecutionPolicy bypass -Force\""
  }
}

build {
  sources = ["sources.googlecompute.windows-ssh-example"]

  provisioner "powershell" {
    script = "../scripts/install-features.ps1"
    elevated_user     = var.packer_username
    elevated_password = var.packer_user_password
  }
}
```

#### Windows over WinSSH - Ansible Provisioner

The following uses Windows SSH as backend communicator
[https://docs.microsoft.com/en-us/windows-server/administration/openssh/openssh_install_firstuse](https://docs.microsoft.com/en-us/windows-server/administration/openssh/openssh_install_firstuse)
with a private key.

* The `sysprep-specialize-script-cmd` creates the `packer_user` and adds it to the local administrators group and configures the ssh key, firewall rule and required permissions.

```
source "googlecompute" "windows-ssh-ansible" {
  project_id              = var.project_id
  source_image            = "windows-server-2019-dc-v20200813"
  zone                    = "us-east4-a"
  disk_size               = 50
  machine_type            = "n1-standard-8"
  communicator            = "ssh"
  ssh_username            = var.packer_username
  ssh_private_key_file    = var.ssh_key_file_path
  ssh_timeout             = "1h"

  metadata = {
    sysprep-specialize-script-cmd = "net user ${var.packer_username} \"${var.packer_user_password}\" /add /y & wmic UserAccount where Name=\"${var.packer_username}\" set PasswordExpires=False & net localgroup administrators ${var.packer_username} /add & powershell Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0 & echo ${var.ssh_pub_key} > C:\\ProgramData\\ssh\\administrators_authorized_keys & icacls.exe \"C:\\ProgramData\\ssh\\administrators_authorized_keys\" /inheritance:r /grant \"Administrators:F\" /grant \"SYSTEM:F\" & powershell New-ItemProperty -Path \"HKLM:\\SOFTWARE\\OpenSSH\" -Name DefaultShell -Value \"C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe\" -PropertyType String -Force  & powershell Start-Service sshd & powershell Set-Service -Name sshd -StartupType 'Automatic' & powershell New-NetFirewallRule -Name 'OpenSSH-Server-In-TCP' -DisplayName 'OpenSSH Server (sshd)' -Enabled True -Direction Inbound -Protocol TCP -Action Allow -LocalPort 22 & powershell.exe -NoProfile -ExecutionPolicy Bypass -Command \"Set-ExecutionPolicy -ExecutionPolicy bypass -Force\""
  }
  account_file = var.account_file_path

}

build {
  sources = ["sources.googlecompute.windows-ssh-ansible"]

  provisioner "ansible" {
    playbook_file           = "./playbooks/playbook.yml"
    use_proxy               = false
    ansible_ssh_extra_args  = ["-o StrictHostKeyChecking=no -o IdentitiesOnly=yes"]
    ssh_authorized_key_file = "var.public_key_path"
    extra_arguments = ["-e", "win_packages=${var.win_packages}",
      "-e",
      "ansible_shell_type=powershell",
      "-e",
      "ansible_shell_executable=None",
      "-e",
      "ansible_shell_executable=None"
    ]
    user = var.packer_username
  }

}

```

#### Nested Hypervisor Example

This is an example of using the `image_licenses` configuration option to create
a GCE image that has nested virtualization enabled. See [Enabling Nested
Virtualization for VM
Instances](https://cloud.google.com/compute/docs/instances/enable-nested-virtualization-vm-instances)
for details.

**JSON**

```json
{
  "builders": [
    {
      "type": "googlecompute",
      "project_id": "my project",
      "source_image_family": "centos-stream-9",
      "ssh_username": "packer",
      "zone": "us-central1-a",
      "image_licenses": ["projects/vm-options/global/licenses/enable-vmx"]
    }
  ]
}
```

**HCL2**

```hcl
source "googlecompute" "basic-example" {
  project_id = "my project"
  source_image_family = "centos-stream-9"
  ssh_username = "packer"
  zone = "us-central1-a"
  image_licenses = ["projects/vm-options/global/licenses/enable-vmx"]
}

build {
  sources = ["sources.googlecompute.basic-example"]
}
```


#### Shared VPC Example

This is an example of using the `network_project_id` configuration option to create
a GCE instance in a Shared VPC Network. See [Creating a GCE Instance using Shared
VPC](https://cloud.google.com/vpc/docs/provisioning-shared-vpc#creating_an_instance_in_a_shared_subnet)
for details. The user/service account running Packer must have `Compute Network User` role on
the Shared VPC Host Project to create the instance in addition to the other roles mentioned in the
Running on Google Cloud section.

**JSON**

```json
{
  "builders": [
    {
      "type": "googlecompute",
      "project_id": "my project",
      "subnetwork": "default",
      "source_image_family": "centos-stream-9",
      "network_project_id": "SHARED_VPC_PROJECT",
      "ssh_username": "packer",
      "zone": "us-central1-a",
      "image_licenses": ["projects/vm-options/global/licenses/enable-vmx"]
    }
  ]
}
```

**HCL2**

```hcl
source "googlecompute" "sharedvpc-example" {
  project_id = "my project"
  source_image_family = "centos-stream-9"
  subnetwork = "default"
  network_project_id = "SHARED_VPC_PROJECT"
  ssh_username = "packer"
  zone = "us-central1-a"
  image_licenses = ["projects/vm-options/global/licenses/enable-vmx"]
}

build {
  sources = ["sources.googlecompute.sharedvpc-example"]
}
```


#### Separate Image Project Example

This is an example of using the `image_project_id` configuration option to create
the generated image in a different GCP project than the one used to create the virtual machine. Make sure that Packer has permission in the target project to manage images, the `Compute Storage Admin` role will grant the desired permissions.

**JSON**

```json
{
  "builders": [
    {
      "type": "googlecompute",
      "project_id": "my project",
      "image_project_id": "my image target project",
      "source_image": "debian-9-stretch-v20200805",
      "ssh_username": "packer",
      "zone": "us-central1-a"
    }
  ]
}
```

**HCL2**

```hcl
source "googlecompute" "basic-example" {
  project_id = "my project"
  image_project_id = "my image target project"
  source_image = "debian-9-stretch-v20200805"
  ssh_username = "packer"
  zone = "us-central1-a"
}

build {
  sources = ["sources.googlecompute.basic-example"]
}
```
