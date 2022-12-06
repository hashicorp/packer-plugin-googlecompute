# Latest Release

Please refer to [releases](https://github.com/hashicorp/packer-plugin-googlecompute/releases) for the latest CHANGELOG information.

---
## 1.0.8 (December 6, 2021)

### Exciting New Features 🎉
* Customer-managed Encryption Key for Remote VM's Boot Disk #20 by @wilsonfv in [#21](https://github.com/hashicorp/packer-plugin-googlecompute/pull/21)
### Doc improvements 📚
* update ansible example by @lmayorga1980 in [#65](https://github.com/hashicorp/packer-plugin-googlecompute/pull/65)

## 1.0.6 (October 18, 2021)

### NOTES:
Support for the HCP Packer registry is currently in beta and requires 
Packer v1.7.7 [GH-47] [GH-52]

### Improvements:
* Add `SourceImageName` as shared builder information variable. [GH-47]
* Add `SourceImageName` to HCP Packer registry image metadata. [GH-47] [GH-52]
* Update Packer plugin SDK to version v0.2.7 [GH-48]

### BUG FIXES:
* Pass DiskName configuration argument when creating instance. [GH-51]

## 1.0.5 (September 13, 2021)

### NOTES:
HCP Packer private beta support requires Packer version 1.7.5 or 1.7.6 [GH-32]

### FEATURES:
* Add HCP Packer registry image metadata to builder artifacts. [GH-32]
* Bump Packer plugin SDK to version v0.2.5 [GH-32]

### IMPROVEMENTS:
* Update driver to use user-configured Service Account for public key import.
    [GH-33]

## 1.0.4 (September 2, 2021)

* Remove Packer core as dependency to plugin. [GH-36]

## 1.0.3 (September 1, 2021)

* Upgrade plugin to use Go 1.17.

## 1.0.2 (August 27, 2021)

* Treat ERROR 4047 as retryable. [GH-34]

## 1.0.0 (June 14, 2021)
The code base for this plugin has been stable since the Packer core split.
We are marking this plugin as v1.0.0 to indicate that it is stable and ready for consumption via `packer init`.

* Update packer-plugin-sdk to v0.2.3
* Update IAP tunnel support to work with all builder authentication types. [GH-19]


## 0.0.2 (April 21, 2021)

* Google Compute plugin break out from Packer core. Changes prior to break out can be found in [Packer's CHANGELOG](https://github.com/hashicorp/packer/blob/master/CHANGELOG.md)
