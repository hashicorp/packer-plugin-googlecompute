// Copyright IBM Corp. 2013, 2026

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

func CreateShieldedVMStateConfig(imagePlatformKey string, imageKeyExchangeKey []string, imageSignaturesDB []string, imageForbiddenSignaturesDB []string) (*compute.InitialStateConfig, error) {
	// When no Secure Boot signature inputs are configured, return nil so the
	// caller leaves ShieldedInstanceInitialState unset on the image payload.
	// Sending an explicit (even empty) InitialStateConfig replaces the
	// PK/KEKs/db/dbx that would otherwise be inherited from the source disk,
	// which causes Secure Boot to fail on VMs launched from the resulting
	// image (UEFI: "Status: Security Violation").
	if imagePlatformKey == "" && len(imageKeyExchangeKey) == 0 && len(imageSignaturesDB) == 0 && len(imageForbiddenSignaturesDB) == 0 {
		return nil, nil
	}

	shieldedVMStateConfig := &compute.InitialStateConfig{}
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
	return shieldedVMStateConfig, nil
}
