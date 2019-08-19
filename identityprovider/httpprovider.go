package identityprovider

import (
	"fmt"

	"github.com/dghubble/sling"
	"github.com/iden3/go-iden3-core/core"
	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"gopkg.in/go-playground/validator.v9"
)

type KeyStore interface{}

type ExportParams struct {
	Passphrase string
}

type ImportParams struct {
	Passphrase string
}

type Identity interface {
	Export(exportFilePath string, exportParams ExportParams) error
	Import(importFilePath string, importParams ImportParams) error
	ID() (*core.ID, error)
	AddClaim(claim *merkletree.Entry) error
	AddClaims(claim []*merkletree.Entry) error
	GenProofClaim(claim *merkletree.Entry) (core.ProofClaim, error)
	GenProofClaims(claims []*merkletree.Entry) ([]core.ProofClaim, error)
	// EmittedClaims() (Array<Claim>, error)
	// ReceivedClaims() (Array<(Claim, ProofClaim)>, error)
	// DataObjects() ([]Data, error)
}

type ServerError struct {
	Err string `json:"error"`
}

func (e ServerError) Error() string {
	return fmt.Sprintf("server: %v", e.Err)
}

type HttpProvider struct {
	Url      string
	Username string
	Password string
	client   *sling.Sling
	validate *validator.Validate
}

type HttpProviderParams struct {
	Url      string
	Username string
	Password string
}

func NewHttpProvider(params HttpProviderParams) *HttpProvider {
	url := params.Url
	if url[len(url)-1] != '/' {
		url += "/"
	}
	client := sling.New().Base(url)
	return &HttpProvider{Url: url, Username: params.Username, Password: params.Password,
		client: client, validate: validator.New()}
}

type HttpIdentity struct {
	Provider *HttpProvider
	ID       core.ID
}

func (p *HttpProvider) request(s *sling.Sling, res interface{}) error {
	var serverError ServerError
	resp, err := s.Receive(res, &serverError)
	if err == nil {
		defer resp.Body.Close()
		if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
			err = serverError
		} else {
			err = p.validate.Struct(res)
		}
	}
	return err
}

type CreateIdentityReq struct {
	ClaimAuthKOp       *merkletree.Entry   `json:"claimAuthKOp" binding:"required"`
	ExtraGenesisClaims []*merkletree.Entry `json:"extraGenesisClaims" binding:"required"`
}

type CreateIdentityRes struct {
	Id *core.ID `json:"id" validate:"required"`
	// ProofOpKey string `json:"proofOpKey"`
}

func (p *HttpProvider) CreateIdentity(keyStore KeyStore, kOp *babyjub.PublicKey,
	extraGenesisClaims []*merkletree.Entry) (*core.ID, error) {

	claimAuthKOp := core.NewClaimAuthorizeKSignBabyJub(kOp)
	createIdentityReq := CreateIdentityReq{
		ClaimAuthKOp:       claimAuthKOp.Entry(),
		ExtraGenesisClaims: extraGenesisClaims,
	}

	var createIdentityRes CreateIdentityRes
	err := p.request(p.client.Path("identity").Post("").BodyJSON(createIdentityReq), &createIdentityRes)
	if err != nil {
		return nil, err
	}

	return createIdentityRes.Id, nil
}

func (p *HttpProvider) LoadIdentity(id *core.ID, keyStore KeyStore) (*HttpIdentity, error) {
	return nil, fmt.Errorf("TODO")
}
