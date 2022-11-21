package oidc

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/radianteam/framework/adapter"
	"github.com/sirupsen/logrus"
)

type OidcConfig struct {
	OfflineMode     bool     `json:"OfflineMode,omitempty" config:"OfflineMode"`
	ProviderUrl     string   `json:"ProviderUrl" config:"ProviderUrl,required"`
	ClientId        string   `json:"ClientId" config:"ClientId,required"`
	ClientSecret    string   `json:"ClientSecret" config:"ClientSecret"`
	RedirectURL     string   `json:"RedirectURL,omitempty" config:"RedirectURL"`
	Scopes          []string `json:"Scopes,omitempty" config:"Scopes"`
	PublicKeys      []string `json:"PublicKeys,omitempty" config:"PublicKeys"`
	SkipIssuerCheck bool     `json:"SkipIssuerCheck,omitempty" config:"SkipIssuerCheck"`
}

type OidcAdapter struct {
	*adapter.BaseAdapter

	config           *OidcConfig
	provider         *oidc.Provider
	staticPublicKeys *oidc.StaticKeySet
}

func NewOidcAdapter(name string, config *OidcConfig) *OidcAdapter {
	return &OidcAdapter{
		BaseAdapter:      adapter.NewBaseAdapter(name),
		config:           config,
		staticPublicKeys: &oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{}}}
}

func (a *OidcAdapter) Setup() (err error) {
	if !a.config.OfflineMode {
		a.provider, err = oidc.NewProvider(context.TODO(), a.config.ProviderUrl)

		if err != nil {
			logrus.WithField("adapter", a.GetName()).Errorf("cannot craete new provider - %s", err)
		}
	}

	if len(a.config.PublicKeys) > 0 {
		for _, pubPEM := range a.config.PublicKeys {
			block, _ := pem.Decode([]byte(pubPEM))

			if block == nil {
				logrus.WithField("adapter", a.GetName()).Errorf("failed to parse PEM block containing the public key")
				continue
			}

			pub, err := x509.ParseCertificate(block.Bytes)

			if err != nil {
				logrus.WithField("adapter", a.GetName()).Errorf("failed to parse DER encoded public key: %s", err)
				continue
			}

			a.staticPublicKeys.PublicKeys = append(a.staticPublicKeys.PublicKeys, pub.PublicKey)
		}
	}

	return
}

func (a *OidcAdapter) Close() (err error) {
	// nothing to do
	return
}

func (a *OidcAdapter) GetVerifier() *oidc.IDTokenVerifier {
	if a.config.OfflineMode {
		return oidc.NewVerifier(a.config.ProviderUrl, a.staticPublicKeys, &oidc.Config{ClientID: a.config.ClientId, SkipIssuerCheck: a.config.SkipIssuerCheck})
	} else if a.provider != nil {
		return a.provider.Verifier(&oidc.Config{ClientID: a.config.ClientId, SkipIssuerCheck: a.config.SkipIssuerCheck})
	}

	logrus.WithField("adapter", a.GetName()).Errorf("cannot get verifier - provider empty")
	return nil
}

// Introspect - remote keycloak function is being called. Before the call, add client_id and client_secret in settings app.
func (a *OidcAdapter) Introspect(token string) (err error) {
	tokenURL := ""

	if a.config.OfflineMode {
		tokenURL = a.config.ProviderUrl + "/protocol/openid-connect/token"
	} else if a.provider != nil {
		tokenURL = a.provider.Endpoint().TokenURL
	} else {
		logrus.WithField("adapter", a.GetName()).Errorf("cannot introspect token - empty provider or offline mode disable")
		return errors.New("empty provider or offline mode disable")
	}

	requestData := make(url.Values)
	requestData.Set("token", token)
	requestData.Set("client_id", a.config.ClientId)
	requestData.Set("client_secret", a.config.ClientSecret)

	client := &http.Client{}
	res, err := client.PostForm(tokenURL+"/introspect", requestData)

	if err != nil {
		logrus.WithField("adapter", a.GetName()).Errorf("cannot introspect token - %s", err)
		return
	}

	defer client.CloseIdleConnections()
	defer res.Body.Close()

	if res.StatusCode != 200 {
		logrus.WithField("adapter", a.GetName()).Errorf("cannot introspect token - %s", res.Status)
		return errors.New(res.Status)
	}

	buf, err := io.ReadAll(res.Body)

	if err != nil {
		logrus.WithField("adapter", a.GetName()).Errorf("cannot read response body - %s", err)
		return
	}

	tokenInfo := make(map[string]any)
	err = json.Unmarshal(buf, &tokenInfo)

	if err != nil {
		logrus.WithField("adapter", a.GetName()).Errorf("cannot unmarshal raw token data - %s", err)
		return
	}

	isActive, ok := tokenInfo["active"]

	if ok && !isActive.(bool) {
		return errors.New("token expired")
	} else {
		expr := time.Unix(int64(tokenInfo["exp"].(float64)), 0)

		if time.Now().UTC().After(expr.UTC()) {
			return errors.New("token expired")
		}
	}

	return
}

func (a *OidcAdapter) VerifyToken(token string) (err error) {
	tokenInfo, err := a.TokenInfo(token)

	if err != nil {
		return
	}

	customStruct := &struct{ Active *bool }{}

	err = tokenInfo.Claims(customStruct)

	if err != nil {
		return errors.New("token claim error")
	}

	if customStruct.Active != nil && !*customStruct.Active || time.Now().UTC().After(tokenInfo.Expiry.UTC()) {
		return errors.New("token expired")
	}

	return
}

func (a *OidcAdapter) TokenInfo(token string) (tokenInfo *oidc.IDToken, err error) {
	verifier := a.GetVerifier()

	if verifier == nil {
		err = errors.New("verifier is empty")
		logrus.WithField("adapter", a.GetName()).Errorf("cannot get token info - %s", err)

		return
	}

	tokenInfo, err = verifier.Verify(context.TODO(), token)

	if err != nil {
		logrus.WithField("adapter", a.GetName()).Errorf("cannot get token info - %s", err)
	}

	return
}
