## 1.0.3 (September 13, 2021)

* Add HCP Packer registry image metadata to builder artifacts. [GH-32]
* Bump Packer plugin SDK to version v0.2.5 [GH-32]
* Update driver to use user-configured Service Account for public key import.
    [GH-33]

## 1.0.2 (August 27, 2021)

* Treat ERROR 4047 as retryable. [GH-34]

## 1.0.0 (June 14, 2021)
The code base for this plugin has been stable since the Packer core split.
We are marking this plugin as v1.0.0 to indicate that it is stable and ready for consumption via `packer init`.

* Update packer-plugin-sdk to v0.2.3
* Update IAP tunnel support to work with all builder authentication types. [GH-19]


## 0.0.2 (April 21, 2021)

* Google Compute plugin break out from Packer core. Changes prior to break out can be found in [Packer's CHANGELOG](https://github.com/hashicorp/packer/blob/master/CHANGELOG.md)
