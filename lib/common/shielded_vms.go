package common

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"

	"google.golang.org/api/compute/v1"
)

func FillFileContentBuffer(certOrKeyFile string) (*compute.FileContentBuffer, error) {
	data, err := os.ReadFile(certOrKeyFile)
	if err != nil {
		err := fmt.Errorf("Unable to read Certificate or Key file %s", certOrKeyFile)
		return nil, err
	}
	shield := &compute.FileContentBuffer{
		Content:  base64.StdEncoding.EncodeToString(data),
		FileType: "X509",
	}
	block, _ := pem.Decode(data)

	if block == nil || block.Type != "CERTIFICATE" {
		_, err = x509.ParseCertificate(data)
	} else {
		_, err = x509.ParseCertificate(block.Bytes)
	}
	if err != nil {
		shield.FileType = "BIN"
	}
	return shield, nil
}

func CreateShieldedVMStateConfig(imageGuestOsFeatures []string, imagePlatformKey string, imageKeyExchangeKey []string, imageSignaturesDB []string, imageForbiddenSignaturesDB []string) (*compute.InitialStateConfig, error) {
	shieldedVMStateConfig := &compute.InitialStateConfig{}
	for _, v := range imageGuestOsFeatures {
		if v == "UEFI_COMPATIBLE" {
			if imagePlatformKey != "" {
				shieldedData, err := FillFileContentBuffer(imagePlatformKey)
				if err != nil {
					return nil, err
				}
				shieldedVMStateConfig.Pk = shieldedData
			}
			for _, v := range imageKeyExchangeKey {
				shieldedData, err := FillFileContentBuffer(v)
				if err != nil {
					return nil, err
				}
				shieldedVMStateConfig.Keks = append(shieldedVMStateConfig.Keks, shieldedData)
			}
			for _, v := range imageSignaturesDB {
				shieldedData, err := FillFileContentBuffer(v)
				if err != nil {
					return nil, err
				}
				shieldedVMStateConfig.Dbs = append(shieldedVMStateConfig.Dbs, shieldedData)
			}
			for _, v := range imageForbiddenSignaturesDB {
				shieldedData, err := FillFileContentBuffer(v)
				if err != nil {
					return nil, err
				}
				shieldedVMStateConfig.Dbxs = append(shieldedVMStateConfig.Dbxs, shieldedData)
			}

		}
	}
	return shieldedVMStateConfig, nil
}
