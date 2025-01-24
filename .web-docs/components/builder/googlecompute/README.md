Type: `googlecompute`
Artifact BuilderId: `packer.googlecompute`

The `googlecompute` Packer builder is able to create
[images](https://developers.google.com/compute/docs/images) for use with
[Google Compute Engine](https://cloud.google.com/products/compute-engine) (GCE)
based on existing images.

It is possible to build images from scratch, but not with the `googlecompute`
Packer builder. The process is recommended only for advanced users, please see
[Building GCE Images from Scratch](https://cloud.google.com/compute/docs/tutorials/building-images)
and the [Google Compute Import
Post-Processor](/packer/integrations/hashicorp/googlecompute/latest/components/post-processor/googlecompute-import) for more
information.

## Authentication

To authenticate with GCE, this builder supports everything the plugin does.
To get more information on this, refer to the plugin's description page, under
the [authentication](/packer/integrations/hashicorp/googlecompute#authentication) section.

## Configuration Reference

Configuration options are organized below into two categories: required and
optional. Within each category, the available options are alphabetized and
described.

In addition to the options listed here, a
[communicator](/packer/docs/templates/legacy_json_templates/communicator) can be configured for this
builder.

### Required:

<!-- Code generated from the comments of the Config struct in builder/googlecompute/config.go; DO NOT EDIT MANUALLY -->

- `project_id` (string) - The project ID that will be used to launch instances and store images.

- `source_image` (string) - The source image to use to create the new image from. You can also
  specify source_image_family instead. If both source_image and
  source_image_family are specified, source_image takes precedence.
  Example: `"debian-8-jessie-v20161027"`

- `source_image_family` (string) - The source image family to use to create the new image from. The image
  family always returns its latest image that is not deprecated. Example:
  `"debian-8"`.

- `zone` (string) - The zone in which to launch the instance used to create the image.
  Example: `"us-central1-a"`

<!-- End of code generated from the comments of the Config struct in builder/googlecompute/config.go; -->


### Optional:

<!-- Code generated from the comments of the Config struct in builder/googlecompute/config.go; DO NOT EDIT MANUALLY -->

- `accelerator_type` (string) - Full or partial URL of the guest accelerator type. GPU accelerators can
  only be used with `"on_host_maintenance": "TERMINATE"` option set.
  Example:
  `"projects/project_id/zones/europe-west1-b/acceleratorTypes/nvidia-tesla-k80"`

- `accelerator_count` (int64) - Number of guest accelerator cards to add to the launched instance.

- `address` (string) - The name of a pre-allocated static external IP address. Note, must be
  the name and not the actual IP address.

- `disable_default_service_account` (bool) - If true, the default service account will not be used if
  service_account_email is not specified. Set this value to true and omit
  service_account_email to provision a VM with no service account.

- `disk_name` (string) - The name of the disk, if unset the instance name will be used.

- `disk_size` (int64) - The size of the disk in GB. This defaults to 20, which is 20GB.

- `disk_type` (string) - Type of disk used to back your instance, like pd-ssd or pd-standard.
  Defaults to pd-standard.

- `disk_encryption_key` (\*common.CustomerEncryptionKey) - Disk encryption key to apply to the created boot disk. Possible values:
  * kmsKeyName -  The name of the encryption key that is stored in Google Cloud KMS.
  * RawKey: - A 256-bit customer-supplied encryption key, encodes in RFC 4648 base64.
  
  examples:
  
   ```json
   {
      "kmsKeyName": "projects/${project}/locations/${region}/keyRings/computeEngine/cryptoKeys/computeEngine/cryptoKeyVersions/4"
   }
   ```
  
   ```hcl
    disk_encryption_key {
      kmsKeyName = "projects/${var.project}/locations/${var.region}/keyRings/computeEngine/cryptoKeys/computeEngine/cryptoKeyVersions/4"
    }
   ```
  
  Refer to the [Customer Encryption Key](#customer-encryption-key) section for more information on the contents of this block.

- `enable_nested_virtualization` (bool) - Create a instance with enabling nested virtualization.

- `enable_secure_boot` (bool) - Create a Shielded VM image with Secure Boot enabled. It helps ensure that
  the system only runs authentic software by verifying the digital signature
  of all boot components, and halting the boot process if signature verification
  fails. [Details](https://cloud.google.com/security/shielded-cloud/shielded-vm)

- `enable_vtpm` (bool) - Create a Shielded VM image with virtual trusted platform module
  Measured Boot enabled. A vTPM is a virtualized trusted platform module,
  which is a specialized computer chip you can use to protect objects,
  like keys and certificates, that you use to authenticate access to your
  system. [Details](https://cloud.google.com/security/shielded-cloud/shielded-vm)

- `enable_integrity_monitoring` (bool) - Integrity monitoring helps you understand and make decisions about the
  state of your VM instances. Note: integrity monitoring relies on having
  vTPM enabled. [Details](https://cloud.google.com/security/shielded-cloud/shielded-vm)

- `disk_attachment` ([]common.BlockDevice) - Extra disks to attach to the instance that will build the final image.
  
  You may reference an existing external persistent disk, or you can configure
  a set of disks to be created before the instance is created, and will
  be deleted upon build completion.
  
  Scratch (ephemeral) SSDs are always created at launch, and deleted when the
  instance is torn-down.
  
  Note: local SSDs are not supported on all machine types, refer to the
  [docs](https://cloud.google.com/compute/docs/disks/local-ssd#choose_number_local_ssds)
  for more information on that.
  
  Refer to the [Extra Disk Attachments](#extra-disk-attachments) section for
  more information on this configuration type.

- `skip_create_image` (bool) - Skip creating the image. Useful for setting to `true` during a build test stage. Defaults to `false`.

- `image_architecture` (string) - The architecture of the resulting image.
  
  Defaults to unset: GCE will use the origin image architecture.

- `image_name` (string) - The unique name of the resulting image. Defaults to
  `packer-{{timestamp}}`.

- `image_description` (string) - The description of the resulting image.

- `image_encryption_key` (\*common.CustomerEncryptionKey) - Image encryption key to apply to the created image. Possible values:
  * kmsKeyName -  The name of the encryption key that is stored in Google Cloud KMS.
  * RawKey: - A 256-bit customer-supplied encryption key, encodes in RFC 4648 base64.
  
  examples:
  
   ```json
   {
      "kmsKeyName": "projects/${project}/locations/${region}/keyRings/computeEngine/cryptoKeys/computeEngine/cryptoKeyVersions/4"
   }
   ```
  
   ```hcl
    image_encryption_key {
      kmsKeyName = "projects/${var.project}/locations/${var.region}/keyRings/computeEngine/cryptoKeys/computeEngine/cryptoKeyVersions/4"
    }
   ```
  
  Refer to the [Customer Encryption Key](#customer-encryption-key) section for more information on the contents of this block.

- `image_family` (string) - The name of the image family to which the resulting image belongs. You
  can create disks by specifying an image family instead of a specific
  image name. The image family always returns its latest image that is not
  deprecated.

- `image_labels` (map[string]string) - Key/value pair labels to apply to the created image.

- `image_licenses` ([]string) - Licenses to apply to the created image.

- `image_guest_os_features` ([]string) - Guest OS features to apply to the created image.

- `image_project_id` (string) - The project ID to push the build image into. Defaults to project_id.

- `image_storage_locations` ([]string) - Storage location, either regional or multi-regional, where snapshot
  content is to be stored and only accepts 1 value. Always defaults to a nearby regional or multi-regional
  location.
  
  multi-regional example:
  
   ```json
   {
      "image_storage_locations": ["us"]
   }
   ```
  regional example:
  
   ```json
   {
      "image_storage_locations": ["us-east1"]
   }
   ```

- `instance_name` (string) - A name to give the launched instance. Beware that this must be unique.
  Defaults to `packer-{{uuid}}`.

- `labels` (map[string]string) - Key/value pair labels to apply to the launched instance.

- `machine_type` (string) - The machine type. Defaults to "e2-standard-2".

- `metadata` (map[string]string) - Metadata applied to the launched instance.
  All metadata configuration values are expected to be of type string.
  Google metadata options that take a value of `TRUE` or `FALSE` should be
  set as a string (i.e  `"TRUE"` `"FALSE"` or `"true"` `"false"`).

- `metadata_files` (map[string]string) - Metadata applied to the launched instance. Values are files.

- `min_cpu_platform` (string) - A Minimum CPU Platform for VM Instance. Availability and default CPU
  platforms vary across zones, based on the hardware available in each GCP
  zone.
  [Details](https://cloud.google.com/compute/docs/instances/specify-min-cpu-platform)

- `network` (string) - The Google Compute network id or URL to use for the launched instance.
  Defaults to "default". If the value is not a URL, it will be
  interpolated to
  `projects/((network_project_id))/global/networks/((network))`. This value
  is not required if a subnet is specified.

- `network_project_id` (string) - The project ID for the network and subnetwork to use for launched
  instance. Defaults to project_id.

- `omit_external_ip` (bool) - If true, the instance will not have an external IP. use_internal_ip must
  be true if this property is true.

- `on_host_maintenance` (string) - Sets Host Maintenance Option. Valid choices are `MIGRATE` and
  `TERMINATE`. Please see [GCE Instance Scheduling
  Options](https://cloud.google.com/compute/docs/instances/setting-instance-scheduling-options),
  as not all machine\_types support `MIGRATE` (i.e. machines with GPUs).
  If preemptible is true this can only be `TERMINATE`. If preemptible is
  false, it defaults to `MIGRATE`

- `preemptible` (bool) - If true, launch a preemptible instance.

- `node_affinity` ([]common.NodeAffinity) - Sets a node affinity label for the launched instance (eg. for sole tenancy).
  Please see [Provisioning VMs on
  sole-tenant nodes](https://cloud.google.com/compute/docs/nodes/provisioning-sole-tenant-vms)
  for more information.
  
  ```hcl
    key = "workload"
    operator = "IN"
    values = ["packer"]
  ```
  
  Refer to the [Node Affinity](#node-affinities) for more information on affinities.

- `state_timeout` (duration string | ex: "1h5m2s") - The time to wait for instance state changes. Defaults to "5m".

- `region` (string) - The region in which to launch the instance. Defaults to the region
  hosting the specified zone.

- `scopes` ([]string) - The service account scopes for launched
  instance. Defaults to:
  
  ```json
  [
    "https://www.googleapis.com/auth/userinfo.email",
    "https://www.googleapis.com/auth/compute",
    "https://www.googleapis.com/auth/devstorage.full_control"
  ]
  ```

- `service_account_email` (string) - The service account to be used for launched instance. Defaults to the
  project's default service account unless disable_default_service_account
  is true.

- `source_image_project_id` ([]string) - A list of project IDs to search for the source image. Packer will search the first
  project ID in the list first, and fall back to the next in the list, until it finds the source image.

- `startup_script_file` (string) - The path to a startup script to run on the launched instance from which the image will
  be made. When set, the contents of the startup script file will be added to the instance metadata
  under the `"startup_script"` metadata property. See [Providing startup script contents directly](https://cloud.google.com/compute/docs/startupscript#providing_startup_script_contents_directly) for more details.
  
  When using `startup_script_file` the following rules apply:
  - The contents of the script file will overwrite the value of the `"startup_script"` metadata property at runtime.
  - The contents of the script file will be wrapped in Packer's startup script wrapper, unless `wrap_startup_script` is disabled. See `wrap_startup_script` for more details.
  - Not supported by Windows instances. See [Startup Scripts for Windows](https://cloud.google.com/compute/docs/startupscript#providing_a_startup_script_for_windows_instances) for more details.

- `windows_password_timeout` (duration string | ex: "1h5m2s") - The time to wait for windows password to be retrieved. Defaults to "3m".

- `wrap_startup_script` (boolean) - For backwards compatibility this option defaults to `"true"` in the future it will default to `"false"`.
  If "true", the contents of `startup_script_file` or `"startup_script"` in the instance metadata
  is wrapped in a Packer specific script that tracks the execution and completion of the provided
  startup script. The wrapper ensures that the builder will not continue until the startup script has been executed.
  - The use of the wrapped script file requires that the user or service account
  running the build has the compute.instance.Metadata role.

- `subnetwork` (string) - The Google Compute subnetwork id or URL to use for the launched
  instance. Only required if the network has been created with custom
  subnetting. Note, the region of the subnetwork must match the region or
  zone in which the VM is launched. If the value is not a URL, it will be
  interpolated to
  `projects/((network_project_id))/regions/((region))/subnetworks/((subnetwork))`

- `tags` ([]string) - Assign network tags to apply firewall rules to VM instance.

- `use_internal_ip` (bool) - If true, use the instance's internal IP instead of its external IP
  during building.

- `use_os_login` (bool) - If true, OSLogin will be used to manage SSH access to the compute instance by
  dynamically importing a temporary SSH key to the Google account's login profile,
  and setting the `enable-oslogin` to `TRUE` in the instance metadata.
  Optionally, `use_os_login` can be used with an existing `ssh_username` and `ssh_private_key_file`
  if a SSH key has already been added to the Google account's login profile - See [Adding SSH Keys](https://cloud.google.com/compute/docs/instances/managing-instance-access#add_oslogin_keys).
  
  SSH keys can be added to an individual user account
  
  ```shell-session
  $ gcloud compute os-login ssh-keys add --key-file=/home/user/.ssh/my-key.pub
  
  $ gcloud compute os-login describe-profile
  PosixAccounts:
  - accountId: <project-id>
   gid: '34567890754'
   homeDirectory: /home/user_example_com
   ...
   primary: true
   uid: '2504818925'
   username: user_example_com
  sshPublicKeys:
   000000000000000000000000000000000000000000000000000000000000000a:
     fingerprint: 000000000000000000000000000000000000000000000000000000000000000a
  ```
  
  Or SSH keys can be added to an associated service account
  ```shell-session
  $ gcloud auth activate-service-account --key-file=<path to service account credentials file (e.g account.json)>
  $ gcloud compute os-login ssh-keys add --key-file=/home/user/.ssh/my-key.pub
  
  $ gcloud compute os-login describe-profile
  PosixAccounts:
  - accountId: <project-id>
   gid: '34567890754'
   homeDirectory: /home/sa_000000000000000000000
   ...
   primary: true
   uid: '2504818925'
   username: sa_000000000000000000000
  sshPublicKeys:
   000000000000000000000000000000000000000000000000000000000000000a:
     fingerprint: 000000000000000000000000000000000000000000000000000000000000000a
  ```

- `wait_to_add_ssh_keys` (duration string | ex: "1h5m2s") - The time to wait between the creation of the instance used to create the image,
  and the addition of SSH configuration, including SSH keys, to that instance.
  The delay is intended to protect packer from anything in the instance boot
  sequence that has potential to disrupt the creation of SSH configuration
  (e.g. SSH user creation, SSH key creation) on the instance.
  Note: All other instance metadata, including startup scripts, are still added to the instance
  during it's creation.
  Example value: `5m`.

<!-- End of code generated from the comments of the Config struct in builder/googlecompute/config.go; -->


<!-- Code generated from the comments of the IAPConfig struct in builder/googlecompute/step_start_tunnel.go; DO NOT EDIT MANUALLY -->

- `use_iap` (bool) - Whether to use an IAP proxy.
  Prerequisites and limitations for using IAP:
  - You must manually enable the IAP API in the Google Cloud console.
  - You must have the gcloud sdk installed on the computer running Packer.
  - If you use a service account, you must add it to project level IAP permissions
    in https://console.cloud.google.com/security/iap. To do so, click
    "project" > "SSH and TCP resources" > "All Tunnel Resources" >
    "Add Member". Then add your service account and choose the role
    "IAP-secured Tunnel User" and add any conditions you may care about.

- `iap_localhost_port` (int) - Which port to connect the local end of the IAM localhost proxy to. If
  left blank, Packer will choose a port for you from available ports.

- `iap_hashbang` (string) - What "hashbang" to use to invoke script that sets up gcloud.
  Default: "/bin/sh"

- `iap_ext` (string) - What file extension to use for script that sets up gcloud.
  Default: ".sh"

- `iap_tunnel_launch_wait` (int) - How long to wait, in seconds, before assuming a tunnel launch was
  successful. Defaults to 30 seconds for SSH or 40 seconds for WinRM.

<!-- End of code generated from the comments of the IAPConfig struct in builder/googlecompute/step_start_tunnel.go; -->


### Startup Scripts

Startup scripts can be a powerful tool for configuring the instance from which
the image is made. The builder will wait for a startup script to terminate. A
startup script can be provided via the `startup_script_file` or
`startup-script` instance creation `metadata` field. Therefore, the build time
will vary depending on the duration of the startup script. If
`startup_script_file` is set, the `startup-script` `metadata` field will be
overwritten. In other words, `startup_script_file` takes precedence.

The builder does check for a pass/fail/error signal from the startup
script by tracking the `startup-script-status` metadata. Packer will check if this key
is set to done and if it not set to done before the timeout, Packer will fail the build.

### Windows
A Windows startup script can only be provided as a metadata field option. The
builder will _not_ wait for a Windows startup script to terminate. You have
to ensure that it finishes before the instance shuts down. For a list of
supported startup script keys refer to [Using startup scripts on Windows](https://cloud.google.com/compute/docs/instances/startup-scripts/windows)

```hcl
metadata = {
  sysprep-specialize-script-cmd = "..."
}
```

### Logging

Startup script logs can be copied to a Google Cloud Storage (GCS) location
specified via the `startup-script-log-dest` instance creation `metadata` field.
The GCS location must be writeable by the service account of the instance that Packer created.

### Communicator Configuration

#### Optional:

<!-- Code generated from the comments of the Config struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `communicator` (string) - Packer currently supports three kinds of communicators:
  
  -   `none` - No communicator will be used. If this is set, most
      provisioners also can't be used.
  
  -   `ssh` - An SSH connection will be established to the machine. This
      is usually the default.
  
  -   `winrm` - A WinRM connection will be established.
  
  In addition to the above, some builders have custom communicators they
  can use. For example, the Docker builder has a "docker" communicator
  that uses `docker exec` and `docker cp` to execute scripts and copy
  files.

- `pause_before_connecting` (duration string | ex: "1h5m2s") - We recommend that you enable SSH or WinRM as the very last step in your
  guest's bootstrap script, but sometimes you may have a race condition
  where you need Packer to wait before attempting to connect to your
  guest.
  
  If you end up in this situation, you can use the template option
  `pause_before_connecting`. By default, there is no pause. For example if
  you set `pause_before_connecting` to `10m` Packer will check whether it
  can connect, as normal. But once a connection attempt is successful, it
  will disconnect and then wait 10 minutes before connecting to the guest
  and beginning provisioning.

<!-- End of code generated from the comments of the Config struct in communicator/config.go; -->


<!-- Code generated from the comments of the SSH struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `ssh_host` (string) - The address to SSH to. This usually is automatically configured by the
  builder.

- `ssh_port` (int) - The port to connect to SSH. This defaults to `22`.

- `ssh_username` (string) - The username to connect to SSH with. Required if using SSH.

- `ssh_password` (string) - A plaintext password to use to authenticate with SSH.

- `ssh_ciphers` ([]string) - This overrides the value of ciphers supported by default by Golang.
  The default value is [
    "aes128-gcm@openssh.com",
    "chacha20-poly1305@openssh.com",
    "aes128-ctr", "aes192-ctr", "aes256-ctr",
  ]
  
  Valid options for ciphers include:
  "aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com",
  "chacha20-poly1305@openssh.com",
  "arcfour256", "arcfour128", "arcfour", "aes128-cbc", "3des-cbc",

- `ssh_clear_authorized_keys` (bool) - If true, Packer will attempt to remove its temporary key from
  `~/.ssh/authorized_keys` and `/root/.ssh/authorized_keys`. This is a
  mostly cosmetic option, since Packer will delete the temporary private
  key from the host system regardless of whether this is set to true
  (unless the user has set the `-debug` flag). Defaults to "false";
  currently only works on guests with `sed` installed.

- `ssh_key_exchange_algorithms` ([]string) - If set, Packer will override the value of key exchange (kex) algorithms
  supported by default by Golang. Acceptable values include:
  "curve25519-sha256@libssh.org", "ecdh-sha2-nistp256",
  "ecdh-sha2-nistp384", "ecdh-sha2-nistp521",
  "diffie-hellman-group14-sha1", and "diffie-hellman-group1-sha1".

- `ssh_certificate_file` (string) - Path to user certificate used to authenticate with SSH.
  The `~` can be used in path and will be expanded to the
  home directory of current user.

- `ssh_pty` (bool) - If `true`, a PTY will be requested for the SSH connection. This defaults
  to `false`.

- `ssh_timeout` (duration string | ex: "1h5m2s") - The time to wait for SSH to become available. Packer uses this to
  determine when the machine has booted so this is usually quite long.
  Example value: `10m`.
  This defaults to `5m`, unless `ssh_handshake_attempts` is set.

- `ssh_disable_agent_forwarding` (bool) - If true, SSH agent forwarding will be disabled. Defaults to `false`.

- `ssh_handshake_attempts` (int) - The number of handshakes to attempt with SSH once it can connect.
  This defaults to `10`, unless a `ssh_timeout` is set.

- `ssh_bastion_host` (string) - A bastion host to use for the actual SSH connection.

- `ssh_bastion_port` (int) - The port of the bastion host. Defaults to `22`.

- `ssh_bastion_agent_auth` (bool) - If `true`, the local SSH agent will be used to authenticate with the
  bastion host. Defaults to `false`.

- `ssh_bastion_username` (string) - The username to connect to the bastion host.

- `ssh_bastion_password` (string) - The password to use to authenticate with the bastion host.

- `ssh_bastion_interactive` (bool) - If `true`, the keyboard-interactive used to authenticate with bastion host.

- `ssh_bastion_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with the
  bastion host. The `~` can be used in path and will be expanded to the
  home directory of current user.

- `ssh_bastion_certificate_file` (string) - Path to user certificate used to authenticate with bastion host.
  The `~` can be used in path and will be expanded to the
  home directory of current user.

- `ssh_file_transfer_method` (string) - `scp` or `sftp` - How to transfer files, Secure copy (default) or SSH
  File Transfer Protocol.
  
  **NOTE**: Guests using Windows with Win32-OpenSSH v9.1.0.0p1-Beta, scp
  (the default protocol for copying data) returns a a non-zero error code since the MOTW
  cannot be set, which cause any file transfer to fail. As a workaround you can override the transfer protocol
  with SFTP instead `ssh_file_transfer_method = "sftp"`.

- `ssh_proxy_host` (string) - A SOCKS proxy host to use for SSH connection

- `ssh_proxy_port` (int) - A port of the SOCKS proxy. Defaults to `1080`.

- `ssh_proxy_username` (string) - The optional username to authenticate with the proxy server.

- `ssh_proxy_password` (string) - The optional password to use to authenticate with the proxy server.

- `ssh_keep_alive_interval` (duration string | ex: "1h5m2s") - How often to send "keep alive" messages to the server. Set to a negative
  value (`-1s`) to disable. Example value: `10s`. Defaults to `5s`.

- `ssh_read_write_timeout` (duration string | ex: "1h5m2s") - The amount of time to wait for a remote command to end. This might be
  useful if, for example, packer hangs on a connection after a reboot.
  Example: `5m`. Disabled by default.

- `ssh_remote_tunnels` ([]string) - 

- `ssh_local_tunnels` ([]string) - 

<!-- End of code generated from the comments of the SSH struct in communicator/config.go; -->


- `ssh_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with SSH.
  The `~` can be used in path and will be expanded to the home directory
  of current user.



### Temporary SSH keypair

<!-- Code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; DO NOT EDIT MANUALLY -->

When no ssh credentials are specified, Packer will generate a temporary SSH
keypair for the instance. You can change the algorithm type and bits
settings.

<!-- End of code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; -->


#### Optional:

<!-- Code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `temporary_key_pair_type` (string) - `dsa` | `ecdsa` | `ed25519` | `rsa` ( the default )
  
  Specifies the type of key to create. The possible values are 'dsa',
  'ecdsa', 'ed25519', or 'rsa'.
  
  NOTE: DSA is deprecated and no longer recognized as secure, please
  consider other alternatives like RSA or ED25519.

- `temporary_key_pair_bits` (int) - Specifies the number of bits in the key to create. For RSA keys, the
  minimum size is 1024 bits and the default is 4096 bits. Generally, 3072
  bits is considered sufficient. DSA keys must be exactly 1024 bits as
  specified by FIPS 186-2. For ECDSA keys, bits determines the key length
  by selecting from one of three elliptic curve sizes: 256, 384 or 521
  bits. Attempting to use bit lengths other than these three values for
  ECDSA keys will fail. Ed25519 keys have a fixed length and bits will be
  ignored.
  
  NOTE: DSA is deprecated and no longer recognized as secure as specified
  by FIPS 186-5, please consider other alternatives like RSA or ED25519.

<!-- End of code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; -->


### Gotchas

CentOS and recent Debian images have root ssh access disabled by default. Set
`ssh_username` to any user, which will be created by packer with sudo access.

The machine type must have a scratch disk, which means you can't use an
`f1-micro` or `g1-small` to build images.

## Extra disk attachments

<!-- Code generated from the comments of the BlockDevice struct in lib/common/block_device.go; DO NOT EDIT MANUALLY -->

BlockDevice is a block device attachement/creation to an instance when building an image.

<!-- End of code generated from the comments of the BlockDevice struct in lib/common/block_device.go; -->


These can be defined using the [disk_attachment](#disk_attachment) block in the configuration.

Note that this is an array, and therefore in HCL2 can be defined as multiple blocks, each
one corresponding to a disk that will be attached to the instance you are booting.

Example:

```hcl
source "googlecompute" "example" {
  # Add whichever is necessary to build the image

  disk_attachment {
    volume_type     = "scratch"
    volume_size     = 375
  }

  disk_attachment {
    volume_type     = "pd-standard"
    volume_size     = 25
    interface_type  = "SCSI"
  }
}
```

### Required:

<!-- Code generated from the comments of the BlockDevice struct in lib/common/block_device.go; DO NOT EDIT MANUALLY -->

- `volume_size` (int) - Size of the volume to request, in gigabytes.
  
  The size specified must be in range of the sizes for the chosen volume type.

- `volume_type` (BlockDeviceType) - The volume type is the type of storage to reserve and attach to the instance being provisioned.
  
  The following values are supported by this builder:
  * scratch: local SSD data, always 375 GiB (default)
  * pd_standard: persistent, HDD-backed disk
  * pd_balanced: persistent, SSD-backed disk
  * pd_ssd: persistent, SSD-backed disk, with extra performance guarantees
  * pd_extreme: persistent, fastest SSD-backed disk, with custom IOPS
  * hyperdisk-balanced: persistent hyperdisk volume, bootable
  * hyperdisk-extreme: persistent hyperdisk volume, optimised for performance
  * hyperdisk-ml: persistent, shareable, hyperdisk volume, highest throughput
  * hyperdisk-throughput: persistent hyperdisk volume with flexible throughput
  
  For details on the different types, refer to: https://cloud.google.com/compute/docs/disks#disk-types
  For more information on hyperdisk volumes, refer to: https://cloud.google.com/compute/docs/disks/hyperdisks#throughput

<!-- End of code generated from the comments of the BlockDevice struct in lib/common/block_device.go; -->


### Optional:

<!-- Code generated from the comments of the BlockDevice struct in lib/common/block_device.go; DO NOT EDIT MANUALLY -->

- `attachment_mode` (string) - How to attach the volume to the instance
  
  Can be either READ_ONLY or READ_WRITE (default).

- `create_image` (bool) - If true, an image will be created for this disk, instead of the boot disk.
  
  This only applies to non-scratch disks, and can only be specified on one disk at a
  time.

- `device_name` (string) - The device name as exposed to the OS in the /dev/disk/by-id/google-* directory
  
  If unspecified, the disk will have a default name in the form
  persistent-disk-x with 'x' being a number assigned by GCE
  
  This field only applies to persistent disks, local SSDs will always
  be exposed as /dev/disk/by-id/google-local-nvme-ssd-x.

- `disk_encryption_key` (CustomerEncryptionKey) - Disk encryption key to apply to the requested disk.
  
  Possible values:
  * kmsKeyName -  The name of the encryption key that is stored in Google Cloud KMS.
  * RawKey: - A 256-bit customer-supplied encryption key, encodes in RFC 4648 base64.
  
  Refer to the [Customer Encryption Key](#customer-encryption-key) section for more information on the contents of this block.

- `disk_name` (string) - Name of the disk to create.
  This only applies to non-scratch disks. If the disk is persistent, and
  not specified, Packer will generate a unique name for the disk.
  
  The name must be 1-63 characters long and comply to the regexp
  '[a-z]([-a-z0-9]*[a-z0-9])?'

- `interface_type` (string) - The interface to use for attaching the disk.
  Can be either NVME or SCSI. Defaults to SCSI.
  
  The available options depend on the type of disk, SEE: https://cloud.google.com/compute/docs/disks/persistent-disks#choose_an_interface

- `iops` (int) - The requested IOPS for the disk.
  
  This is only available for pd_extreme disks.

- `keep_device` (bool) - Keep the device in the created disks after the instance is terminated.
  By default, the builder will remove the disks at the end of the build.
  
  This cannot be used with 'scratch' volumes.

- `replica_zones` ([]string) - The list of extra zones to replicate the disk into
  
  The zone in which the instance is created will automatically be
  added to the zones in which the disk is replicated.

- `source_volume` (string) - The URI of the volume to attach
  
  If this is specified, it won't be deleted after the instance is shut-down.

- `source_image` (string) - The URI of the image to load
  
  This cannot be used with SourceVolume.

- `_` (string) - Zone is the zone in which to create the disk in.
  
  It is not exposed since the parent config already specifies it
  and it will be set for the block device when preparing it.

<!-- End of code generated from the comments of the BlockDevice struct in lib/common/block_device.go; -->


## Customer Encryption Key

Specifying a custom key allows you to use your own encryption keys to encrypt the data
of the image you are creating.

Note: you will need to reuse the same key later on when reusing the image.

<!-- Code generated from the comments of the CustomerEncryptionKey struct in lib/common/client_keys.go; DO NOT EDIT MANUALLY -->

- `kmsKeyName` (string) - KmsKeyName: The name of the encryption key that is stored in Google
  Cloud KMS.

- `rawKey` (string) - RawKey: Specifies a 256-bit customer-supplied encryption key, encoded
  in RFC 4648 base64 to either encrypt or decrypt this resource.

<!-- End of code generated from the comments of the CustomerEncryptionKey struct in lib/common/client_keys.go; -->


## Node Affinities

Node affinity configuration allows you to restrict the nodes on which to run the
instance that Packer will build the image from.
This requires configuring [sole-tenant node groups](https://cloud.google.com/compute/docs/nodes/provisioning-sole-tenant-vms) first.

<!-- Code generated from the comments of the NodeAffinity struct in lib/common/affinities.go; DO NOT EDIT MANUALLY -->

- `key` (string) - Key: Corresponds to the label key of Node resource.

- `operator` (string) - Operator: Defines the operation of node selection. Valid operators are IN for affinity and
  NOT_IN for anti-affinity.

- `values` ([]string) - Values: Corresponds to the label values of Node resource.

<!-- End of code generated from the comments of the NodeAffinity struct in lib/common/affinities.go; -->
