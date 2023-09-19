// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googlecompute

import (
	"os"
	"path"
	"testing"
)

var (
	JWT_JSON = []byte(`{
		"type": "service_account",
		"project_id": "myproject",
		"private_key_id": "13b418866f17f421c8c2040f4c0eeca6da48d8c1",
		"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC2QZwjhFbHC68C\nJYUOrCqy13KFg0HGl9KiRiAKuGN4beu7D4vuFLqbEPVik0h6kY+CKkYCyagSoIQE\nwW9/QlGos3HVFFHrEmIK07Yx2KTUUkoiL4LqqOJgLqKeHjCB9G3k33cXk9QthcX9\nzLcih9Jou251uaA0T4iKSlr/aDS64DvQvkqfAXjdp3l/ZJr2NOTJckoo/nsyHF3Y\nPocR8uFPxtzk4IrdFqh63d20qCepZyNFdQ04GJAefdPe5ldLCmAkLgIi0z66iQJd\nz7vm9Ml11EVO+lfhp/iL/vVPPvqK4j7IBw66q4UgzcRXEAcZ5KHw1YvMly0gKzXG\nuUjIuwT1AgMBAAECggEAD9CpeSy3XP+6C+OJysHZokaiXn++M7q77khLtbJvZNiU\nL1mRHABt4VChtanJjP0+XASwEF2p7HsQsA4nzzEUv4dZ76wBbPaqgRcC+D6JrkG5\nXIMNhxa1lU2mrDYH1PCkn2SQXU1LvAS5YnJWF+evGr1DWT4M1U04SdiO64FoeSi5\nyyFZFEkbLFpD57AH/0cBKuWmbqDBdJu1dkjvZM5EFu6C6kH294dfiHTawGpf6Vp3\nt6bFaP8VRAev3U+sJKeR3/1aHomeSIbRaUMeq9rvczKb7j0luyW8wdpre8sIfMZD\no7jCu8aoQWXtflO49mObLU2ohpxsYb0MtoMJBAJ9pwKBgQDuvXwWNO98OyS/S3TB\nmLnvNL5w36JMv7sVo1QiaGhc3UXWnCLIukvjJ6GJd9zGFGslk67o0SAXqoCn6TVH\n32X+eq7aumcioa1Lq3ax09r4mdZgND59cjjiWJvZ8EvU0hZcebbqSw+7CScOZGCj\nddaFUAFpHgZzHy1Eo2fYHywBYwKBgQDDbr4C12xTh9Gv1n8BGss+Ky4rMZVUZGZU\nZIrZKBrJ3JVjN02VH/2Dpn0DBXyoDGBCpza8vvaRbDeEUROQ03Tyc9moXW2CQPjC\nt33g9DAwPeA2/GPCQfAdeKB6zuTo5XDSL7ho6GLdmF/dhyH+yXy0GeyE6YCSHNZY\nAnj0I0WbxwKBgDBGWIUVBygTvYaA94b+Hvrjq26fie4DBw2FDUo32oKMq8aNo+r6\n4MV6CgwGFLpo/pGGn2OshdTDQWiym3eBENq4bAsGjjxOfQBEF6g1sp16XgLuDYTI\nSABc8obLNEpAgQ0J/5a4vuGPJDqgyXnEJjCm0OI0lBFLSJgMgr8M7pUJAoGAUnY2\n3LITNke31Y8XNdsdaRUFPRqF3P8kInXuFGUUsJpPunaKWOMPsG4ej5jQGYRnVZiC\nwy98kK3t2vnu3Iws62SwsZcCbxSFInwUNEg00RY6tljWqw/xhi3w4QDNm+u8KCQU\nlsd/d+skgC/Vy1EvOjs6DncMVhqu4qHgcXs0kt8CgYEAz1TafVCcMY4hISPwKKH/\nsdwxRTt1/88yWz7KgtNRWzRJMTyjBCV0eRZJx1NhGwEIkmNgpeg87CN0GDdd0Nif\npC/9gEnRDlfk8MxYdkdMUUi+QZ43dvem6SsedRnnIRGKlNt2zU8TJEENmsSM+nbP\n2zVjYfO2qsmV2pMvx9eEdBo=\n-----END PRIVATE KEY-----\n",
		"client_email": "dummy@myproject.iam.gserviceaccount.com",
		"client_id": "123456789",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/myproject%40myproject.iam.gserviceaccount.com",
		"universe_domain": "googleapis.com"
	  }`)
	OIDC_JSON = []byte(`{
		"type":"external_account",
	"audience":"//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/poolName/providers/some-provider-name",
	"subject_token_type":"urn:ietf:params:oauth:token-type:jwt","token_url":"https://sts.googleapis.com/v1/token",
	"service_account_impersonation_url":"https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/some-service-account@some-gcp-project.iam.gserviceaccount.com:generateAccessToken",
	"credential_source":{"url":"https://pipelines.actions.githubusercontent.com/blahblahblah",
	"headers":{"Authorization":"***"},
	"format":{"type":"json","subject_token_field_name":"value"}}}
	`)
)

func TestProcessAccountFile_JWTtext(t *testing.T) {
	account, err := ProcessAccountFile(string(JWT_JSON))
	if err != nil {
		t.Errorf("Bad: unexpected error when trying to load JWT credential %v", err)
	}
	if account == nil {
		t.Errorf("Bad: Account file could not be loaded")
	}
	if account != nil && account.jwt == nil {
		t.Errorf("Bad: JWT credential was nil")
	}
	if account.jwt.Email != "dummy@myproject.iam.gserviceaccount.com" {
		t.Errorf("Bad: JWT credential email was %s, expected 'dummy@myproject.iam.gserviceaccount.com'", account.jwt.Email)
	}
}

func TestProcessAccountFile_JWTFile(t *testing.T) {
	workdir := t.TempDir()
	credsFile := path.Join(workdir, "jwt-file")
	err := os.WriteFile(credsFile, JWT_JSON, 0777)
	if err != nil {
		t.Errorf("Bad: error writing temporary JWT creds file: %v", err)
	}
	account, err := ProcessAccountFile(credsFile)
	if err != nil {
		t.Errorf("Bad: unexpected error when trying to load JWT credential: %v", err)
	}
	if account == nil {
		t.Errorf("Bad: Account file could not be loaded")
	}
	if account != nil && account.jwt == nil {
		t.Errorf("Bad: JWT credential was nil")
	}
	if account.jwt.Email != "dummy@myproject.iam.gserviceaccount.com" {
		t.Errorf("Bad: JWT credential email was %s, expected 'dummy@myproject.iam.gserviceaccount.com'", account.jwt.Email)
	}
}

func TestProcessAccountFile_OIDC(t *testing.T) {
	account, err := ProcessAccountFile(string(OIDC_JSON))
	if err != nil {
		t.Errorf("Bad: unexpected error when trying to load OIDC credential: %v", err)
	}
	if account == nil {
		t.Errorf("Bad: Account file could not be loaded")
	}
	if account != nil && account.jwt != nil {
		t.Errorf("Bad: JWT credential was not nil (should be nil for OIDC account file)")
	}
}
