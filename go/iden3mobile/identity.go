package iden3mobile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/components/httpclient"
	"github.com/iden3/go-iden3-core/components/idenpuboffchain/readerhttp"
	"github.com/iden3/go-iden3-core/db"
	"github.com/iden3/go-iden3-core/identity/holder"
	"github.com/iden3/go-iden3-crypto/babyjub"
	issuerMsg "github.com/iden3/go-iden3-servers-demo/servers/issuerdemo/messages"
	verifierMsg "github.com/iden3/go-iden3-servers-demo/servers/verifier/messages"
	log "github.com/sirupsen/logrus"
)

type Callback interface {
	VerifierResHandler(bool, error)
	RequestClaimResHandler(*Ticket, error)
}

type Identity struct {
	id          *holder.Holder
	storage     db.Storage
	ClaimDB     *ClaimDB
	Tickets     *Tickets
	stopTickets chan bool
	eventQueue  chan Event
}

const (
	kOpStorKey           = "kOpComp"
	storageSubPath       = "/idStore"
	keyStorageSubPath    = "/idKeyStore"
	smartContractAddress = "0xF6a014Ac66bcdc1BF51ac0fa68DF3f17f4b3e574"
	credExisPrefix       = "credExis"
)

func isEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// NewIdentity creates a new identity
// this funciton is mapped as a constructor in Java.
// NOTE: The storePath must be unique per Identity.
// NOTE: Right now the extraGenesisClaims is useless.
func NewIdentity(storePath, pass, web3Url string, checkTicketsPeriodMilis int, extraGenesisClaims *BytesArray) (*Identity, error) {
	// Check that storePath points to an empty dir
	if dirIsEmpty, err := isEmpty(storePath); !dirIsEmpty || err != nil {
		if err == nil {
			err = errors.New("Directory is not empty")
		}
		return nil, err
	}
	idenPubOnChain, keyStore, storage, err := loadComponents(storePath, web3Url)
	if err != nil {
		return nil, err
	}
	resourcesAreClosed := false
	defer func() {
		if !resourcesAreClosed {
			keyStore.Close()
			storage.Close()
		}
	}()
	// Create babyjub keys
	kOpComp, err := keyStore.NewKey([]byte(pass))
	if err != nil {
		return nil, err
	}
	if err = keyStore.UnlockKey(kOpComp, []byte(pass)); err != nil {
		return nil, err
	}
	// Store kOpComp
	tx, err := storage.NewTx()
	if err != nil {
		return nil, err
	}
	if err := db.StoreJSON(tx, []byte(kOpStorKey), kOpComp); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	// TODO: Parse extra genesis claims. Call toClaimers once it's implemented
	// _extraGenesisClaims, err := extraGenesisClaims.toClaimers()
	if err != nil {
		return nil, err
	}
	// Create new Identity (holder)
	_, err = holder.New(
		holder.ConfigDefault,
		kOpComp,
		nil,
		storage,
		keyStore,
		idenPubOnChain,
		nil,
		readerhttp.NewIdenPubOffChainHttp(),
	)
	if err != nil {
		return nil, err
	}
	// Init claim DB
	// cdb := NewClaimDB(storage.WithPrefix([]byte(credExisPrefix)))
	// Verify that the Identity can be loaded successfully
	keyStore.Close()
	storage.Close()
	resourcesAreClosed = true
	return NewIdentityLoad(storePath, pass, web3Url, checkTicketsPeriodMilis)
}

// NewIdentityLoad loads an already created identity
// this funciton is mapped as a constructor in Java
func NewIdentityLoad(storePath, pass, web3Url string, checkTicketsPeriodMilis int) (*Identity, error) {
	// TODO: figure out how to diferentiate the two constructors from Java: https://github.com/iden3/iden3-mobile/issues/17#issuecomment-587374644
	idenPubOnChain, keyStore, storage, err := loadComponents(storePath, web3Url)
	if err != nil {
		return nil, err
	}
	defer keyStore.Close()
	// Unlock key store
	kOpComp := &babyjub.PublicKeyComp{}
	if err := db.LoadJSON(storage, []byte(kOpStorKey), kOpComp); err != nil {
		return nil, err
	}
	if err := keyStore.UnlockKey(kOpComp, []byte(pass)); err != nil {
		return nil, fmt.Errorf("Error unlocking babyjub key from keystore: %w", err)
	}
	// Load existing Identity (holder)
	holdr, err := holder.Load(storage, keyStore, idenPubOnChain, nil, readerhttp.NewIdenPubOffChainHttp())
	if err != nil {
		return nil, err
	}
	// Init Identity
	iden := &Identity{
		id:          holdr,
		storage:     storage,
		Tickets:     NewTickets(storage.WithPrefix([]byte(ticketPrefix))),
		stopTickets: make(chan bool),
		eventQueue:  make(chan Event, 10),
		ClaimDB:     NewClaimDB(storage.WithPrefix([]byte(credExisPrefix))),
	}
	go iden.Tickets.CheckPending(iden, iden.eventQueue, time.Duration(checkTicketsPeriodMilis)*time.Millisecond, iden.stopTickets)
	return iden, nil
}

// GetNextEvent returns the oldest event that has been generated.
// Note that each event can only be retireved once.
// Note that this function is blocking and potentially for a very long time.
func (i *Identity) GetNextEvent() *Event {
	ev := <-i.eventQueue
	return &ev
}

// Stop close all the open resources of the Identity
func (i *Identity) Stop() {
	log.Info("Stopping identity: ", i.id.ID())
	defer i.storage.Close()
	i.stopTickets <- true
}

// RequestClaim sends a petition to issue a claim to an issuer.
// This function will eventually trigger an event,
// the returned ticket can be used to reference the event
func (i *Identity) RequestClaim(baseUrl, data string, c Callback) {
	go func() {
		id := uuid.New().String()
		t := &Ticket{
			Id:     id,
			Type:   TicketTypeClaimStatus,
			Status: TicketStatusPending,
		}
		httpClient := httpclient.NewHttpClient(baseUrl)
		res := issuerMsg.ResClaimRequest{}
		if err := httpClient.DoRequest(httpClient.NewRequest().Path(
			"claim/request").Post("").BodyJSON(&issuerMsg.ReqClaimRequest{
			Value: data,
		}), &res); err != nil {
			c.RequestClaimResHandler(nil, err)
			return
		}
		t.handler = &reqClaimStatusHandler{
			Id:      res.Id,
			BaseUrl: baseUrl,
		}
		err := i.Tickets.Add([]Ticket{*t})
		c.RequestClaimResHandler(t, err)
	}()
}

// ProveCredential sends a credentialValidity build from the given credentialExistance to a verifier
// the callback is used to check if the verifier has accepted the credential as valid
func (i *Identity) ProveClaim(baseUrl string, credId []byte, c Callback) {
	// TODO: add context
	go func() {
		// Get credential existance
		credExis, err := i.ClaimDB.GetReceivedCredential(credId)
		if err != nil {
			c.VerifierResHandler(false, err)
			return
		}
		// Build credential validity
		credVal, err := i.id.HolderGetCredentialValidity(credExis)
		if err != nil {
			c.VerifierResHandler(false, err)
			return
		}
		// Send credential to verifier
		httpClient := httpclient.NewHttpClient(baseUrl)
		if err := httpClient.DoRequest(httpClient.NewRequest().Path(
			"verify").Post("").BodyJSON(verifierMsg.ReqVerify{
			CredentialValidity: credVal,
		}), nil); err != nil {
			// Credential declined / error
			c.VerifierResHandler(false, err)
			return
		}
		// Success
		c.VerifierResHandler(true, nil)
	}()
}