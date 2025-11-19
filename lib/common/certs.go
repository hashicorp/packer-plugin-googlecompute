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
