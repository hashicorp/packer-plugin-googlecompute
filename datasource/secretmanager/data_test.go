package secretmanager

import (
	"context"
	_ "embed"
	"hash/crc32"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"google.golang.org/grpc"
)

type SecretManagerMock struct {
	secretmanagerpb.UnimplementedSecretManagerServiceServer
}

func (s *SecretManagerMock) Run() (string, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	gsrv := grpc.NewServer()
	secretmanagerpb.RegisterSecretManagerServiceServer(gsrv, s)
	go func() {
		if err := gsrv.Serve(l); err != nil {
			panic(err)
		}
	}()
	return l.Addr().String(), nil
}

func (s *SecretManagerMock) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	payload := []byte("secret_content")
	checksum := int64(crc32.Checksum(payload, crc32.IEEETable))

	return &secretmanagerpb.AccessSecretVersionResponse{
		Name: req.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data:       payload,
			DataCrc32C: &checksum,
		},
	}, nil
}

func TestSecretManagerDataSource_Mock(t *testing.T) {
	m := &SecretManagerMock{}
	addr, err := m.Run()
	if err != nil {
		t.Fatal(err)
	}

	type TestCase struct {
		name            string
		project         string
		version         string
		shouldConfigErr bool
		shouldExecErr   bool
	}

	for name, testcase := range map[string]TestCase{
		"empty": {
			shouldConfigErr: true,
		},
		"empty_name": {
			project:         "prj-test",
			shouldConfigErr: true,
		},
		"empty_project": {
			name:            "test",
			shouldConfigErr: true,
		},
		"empty_version": {
			name:            "test",
			project:         "prj-test",
			shouldConfigErr: false,
		},
	} {
		testcase := testcase
		t.Run(name, func(t *testing.T) {
			d := &Datasource{
				config: Config{
					MockOption: []interface{}{
						option.WithEndpoint(addr),
						option.WithoutAuthentication(),
						option.WithGRPCDialOption(grpc.WithInsecure()),
					},
					Name:    testcase.name,
					Project: testcase.project,
					Version: testcase.version,
				},
			}
			err := d.Configure()
			require.Equal(t, testcase.shouldConfigErr, err != nil)

			if err == nil {
				_, err := d.Execute()
				require.Nil(t, err)
			}
		})
	}
}
