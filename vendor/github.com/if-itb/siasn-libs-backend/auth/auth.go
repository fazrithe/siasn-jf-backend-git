// Package auth provides authentication mechanism with BKN OIDC provider.
// Multiple error codes and messages in this package can be overridden to follow standard error codes and messages per
// service.
//
// This auth package is NOT compatible with siasn-jf-backend auth package. It is the revision of that package.
package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/if-itb/siasn-libs-backend/ec"
	"github.com/if-itb/siasn-libs-backend/httputil"
	"github.com/if-itb/siasn-libs-backend/logutil"
	"github.com/if-itb/siasn-libs-backend/metricutil"
	"golang.org/x/oauth2"
	"gopkg.in/square/go-jose.v2/jwt"
)

const (
	userContextKey       = "oidc-user"
	userDetailContextKey = "user"
)

var (
	ErrCodeNoOAuth2ExchangeCode      = 99401
	ErrMessageNoOAuth2ExchangeCode   = "no OAuth2 exchange code"
	ErrCodeNoIdToken                 = 99402
	ErrMessageNoIdToken              = "no ID token present in request"
	ErrCodeNoWorkAgencyId            = 99403
	ErrMessageNoWorkAgencyId         = "no work agency ID"
	ErrCodeInvalidUser               = 99404
	ErrMessageInvalidUser            = "user detail cannot be found"
	ErrCodeNoUserInSessionStorage    = 99405
	ErrMessageNoUserInSessionStorage = "user does not have currently active session"
	ErrCodeNotAsn                    = 99406
	ErrMessageNotAsn                 = "user is not found in ASN profile DB"
	ErrCodeLoginRequired             = 99407
	ErrMessageLoginRequired          = "prompt is set to none but login is required, redirect the user to the login page to login"
	ErrCodeGenericOAuth2             = 99408
	ErrMessageGenericOAuth2          = "generic error"
	ErrCodeOAuth2ExchangeFailed      = 99501
	ErrMessageOAuth2ExchangeFailed   = "code exchange failed"
	ErrCodeCannotVerifyIdToken       = 99502
	ErrMessageCannotVerifyIdToken    = "cannot verify newly issued ID token"
	ErrCodeProfileQueryFail          = 99503
	ErrMessageProfileQueryFail       = "SQL query fail"
	ErrCodeNonceGenerateFail         = 99504
	ErrMessageNonceGenerateFail      = "unable to generate nonce"
	ErrCodeStateGenerateFail         = 99505
	ErrMessageStateGenerateFail      = "unable to generate state"
)

type LoginChecker interface {
	// CheckLogin checks whether login can proceed.
	// It is the user responsibility to write error and HTTP status code to the client.
	// If user is not allowed, either due to server error or not, error must be non-nil.
	CheckLogin(writer http.ResponseWriter, idToken *oidc.IDToken, user *Asn) (err error)
}

// NoopLoginChecker allows all kinds of login.
type NoopLoginChecker struct{}

func (n *NoopLoginChecker) CheckLogin(writer http.ResponseWriter, idToken *oidc.IDToken, user *Asn) (err error) {
	return nil
}

// Auth is an authentication handler.
// It requires an OpenID Connect provider.
type Auth struct {
	Logger             logutil.Logger
	config             *oauth2.Config
	provider           *oidc.Provider
	verifier           *oidc.IDTokenVerifier
	SuccessRedirectUrl string
	EndSessionEndpoint string
	ProfileDb          *sql.DB
	ReferenceDb        *sql.DB
	SqlMetrics         metricutil.GenericSqlMetrics
	AccessTokenCache   AccessTokenCache
	NonceCache         NonceCache
	StateCache         StateCache
	LoginChecker       LoginChecker
	CookieSameSite     http.SameSite
	CookieSecure       bool
	CookieHttpOnly     bool
	CookieDomain       string
	AccessTokenRole    string
}

type User struct {
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	Expired           int64  `json:"exp"`
	Subject           string `json:"sub"`
	// AccessToken is retrieved usually from cache, and can be nil.
	AccessToken *AccessToken `json:"-"`
}

// Asn (ASN or PNS) represents a single ASN profile.
type Asn struct {
	// The ASN ID is a 36 uppercase hexadecimal characters string.
	AsnId                       string `json:"asn_id"`
	NewNip                      string `json:"nip_baru"`
	OldNip                      string `json:"nip_lama"`
	Name                        string `json:"nama"`
	Nik                         string `json:"nik"`
	Email                       string `json:"email"`
	PhoneNumber                 string `json:"no_hp"`
	Birthday                    string `json:"tgl_lahir"`
	Username                    string `json:"username"`
	ParentAgencyId              string `json:"instansi_induk_id"`
	ParentAgency                string `json:"instansi_induk_nama"`
	WorkAgencyId                string `json:"instansi_kerja_id"`
	WorkAgency                  string `json:"instansi_kerja_nama"`
	FunctionalPositionId        string `json:"jabatan_fungsional_id"`
	GenericFunctionalPositionId string `json:"jabatan_fungsional_umum_id"`
	Position                    string `json:"jabatan"`
	OrganizationUnit            string `json:"unit_organisasi"`
	BracketId                   string `json:"golongan_id"`
	Bracket                     string `json:"golongan"`
	Rank                        string `json:"pangkat"`
	// AccessToken is retrieved usually from cache, and cannot be nil as to retrieve user detail, access token is
	// required. See also UserDetailAuthHandler.
	AccessToken *AccessToken `json:"-"`
	AccessRole  string       `json:role`
}

func NewAuth(
	providerUrl string,
	profileDb, referenceDb *sql.DB,
	clientId, clientSecret, endSessionEndpoint, redirectUrl, successRedirectUrl string,
) (*Auth, error) {
	provider, err := oidc.NewProvider(context.Background(), providerUrl)
	if err != nil {
		return nil, err
	}

	_, err = url.Parse(endSessionEndpoint)
	if err != nil {
		return nil, err
	}

	config := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectUrl,
		Scopes:       []string{"openid"},
	}

	return &Auth{
		Logger:             logutil.NewStdLogger(true, "auth"),
		provider:           provider,
		config:             config,
		verifier:           provider.Verifier(&oidc.Config{ClientID: clientId}),
		SuccessRedirectUrl: successRedirectUrl,
		EndSessionEndpoint: endSessionEndpoint,
		ProfileDb:          profileDb,
		ReferenceDb:        referenceDb,
		LoginChecker:       &NoopLoginChecker{},
		CookieSameSite:     http.SameSiteLaxMode,
		CookieHttpOnly:     false,
	}, nil
}

// GetUserDetail retrieves user extended detail from the database, based on the user NIP. workAgencyId is optional
// This function does not make distinction about new or old NIP. When the user cannot be found, it will return no error
// but will return nil user instead.
func (a *Auth) GetUserDetail(ctx context.Context, nip string, workAgencyId string) (user *Asn, err error) {
	profileMdb := metricutil.NewDB(a.ProfileDb, a.SqlMetrics)
	referenceMdb := metricutil.NewDB(a.ReferenceDb, a.SqlMetrics)

	user = &Asn{}
	unorId := ""
	positionTypeId := 0
	birthday := sql.NullString{}

	err = profileMdb.QueryRowContext(
		ctx,
		`
select 
       pns.id,
       nip_baru,
       coalesce(nip_lama, ''),
       coalesce(nama, ''),
       coalesce(nomor_id_document, ''),
       coalesce(nomor_hp, ''),
       tgl_lhr,
       instansi_induk_id,
       coalesce(instansi_induk_nama, ''),
       instansi_kerja_id,
       coalesce(instansi_kerja_nama, ''),
       coalesce(jabatan_fungsional_id, ''),
       coalesce(jabatan_fungsional_umum_id, ''),
       jenis_jabatan_id,
       coalesce(unor_id, ''),
       golongan_id
from pns 
    left join orang on pns.id = orang.id
where (nip_baru = $1 or nip_lama = $1) and ($2::text is null or $2 = '' or instansi_kerja_id = $2) order by nip_baru limit 1`,
		nip,
		workAgencyId,
	).Scan(
		&user.AsnId,
		&user.NewNip,
		&user.OldNip,
		&user.Name,
		&user.Nik,
		&user.PhoneNumber,
		&birthday,
		&user.ParentAgencyId,
		&user.ParentAgency,
		&user.WorkAgencyId,
		&user.WorkAgency,
		&user.FunctionalPositionId,
		&user.GenericFunctionalPositionId,
		&positionTypeId,
		&unorId,
		&user.BracketId,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, ec.NewError(ErrCodeProfileQueryFail, ErrMessageProfileQueryFail, err)
	}

	if err == sql.ErrNoRows {
		err = profileMdb.QueryRowContext(
			ctx,
			`
select 
       pppk.id,
       nip_baru,
       coalesce(nama, ''),
       coalesce(nomor_hp, ''),
       coalesce(nomor_id_document, ''),
       instansi_induk_id,
       coalesce(instansi_induk_nama, ''),
       instansi_kerja_id,
       coalesce(instansi_kerja_nama, ''),
       coalesce(jabatan_fungsional_id, ''),
       coalesce(jabatan_fungsional_umum_id, ''),
       jenis_jabatan_id,
       coalesce(unor_id, ''),
       golongan_id
from pppk 
    left join orang on pppk.id = orang.id
where nip_baru = $1 and ($2::text is null or $2 = '' or instansi_kerja_id = $2) order by nip_baru limit 1`,
			nip,
			workAgencyId,
		).Scan(
			&user.AsnId,
			&user.NewNip,
			&user.Name,
			&user.Nik,
			&user.PhoneNumber,
			&birthday,
			&user.ParentAgencyId,
			&user.ParentAgency,
			&user.WorkAgencyId,
			&user.WorkAgency,
			&user.FunctionalPositionId,
			&user.GenericFunctionalPositionId,
			&positionTypeId,
			&unorId,
			&user.BracketId,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, ec.NewError(ErrCodeProfileQueryFail, ErrMessageProfileQueryFail, err)
		}
	}

	role := a.AccessTokenRole
	user.Birthday = birthday.String
	user.AccessRole = role
	unorPosition := ""
	err = referenceMdb.QueryRowContext(ctx, `select nama_unor, coalesce(nama_jabatan, '') from unor where id = $1`, unorId).Scan(&user.OrganizationUnit, &unorPosition)
	if err != nil && err != sql.ErrNoRows {
		return nil, ec.NewError(ErrCodeProfileQueryFail, ErrMessageProfileQueryFail, err)
	}

	if positionTypeId == 1 || positionTypeId == 3 {
		user.Position = unorPosition
	} else if positionTypeId == 2 && user.FunctionalPositionId != "" {
		err = referenceMdb.QueryRowContext(ctx, `select nama from jabatan_fungsional where id = $1`, user.FunctionalPositionId).Scan(&user.Position)
		if err != nil && err != sql.ErrNoRows {
			return nil, ec.NewError(ErrCodeProfileQueryFail, ErrMessageProfileQueryFail, err)
		}
	} else if positionTypeId == 4 && user.GenericFunctionalPositionId != "" {
		err = referenceMdb.QueryRowContext(ctx, `select nama from jabatan_fungsional_umum where id = $1`, user.GenericFunctionalPositionId).Scan(&user.Position)
		if err != nil && err != sql.ErrNoRows {
			return nil, ec.NewError(ErrCodeProfileQueryFail, ErrMessageProfileQueryFail, err)
		}
	}

	err = nil

	err = referenceMdb.QueryRowContext(ctx, "select nama, nama_pangkat from golongan where id = $1", user.BracketId).Scan(&user.Bracket, &user.Rank)
	if err != nil && err != sql.ErrNoRows {
		return nil, ec.NewError(ErrCodeProfileQueryFail, ErrMessageProfileQueryFail, err)
	}

	err = nil

	return user, nil
}

// CreateAuthUrl create an authentication URL.
// The user can be redirected to the auth URL when they are unauthenticated.
func (a *Auth) CreateAuthUrl(prompt, state string) (url string, err error) {
	params := make([]oauth2.AuthCodeOption, 0)
	if a.NonceCache != nil {
		nonce, err := a.NonceCache.GenerateNonce()
		if err != nil {
			return "", ec.NewError(ErrCodeNonceGenerateFail, ErrMessageNonceGenerateFail, err)
		}
		params = append(params, oauth2.SetAuthURLParam("nonce", nonce))
	}

	if prompt != "" {
		params = append(params, oauth2.SetAuthURLParam("prompt", prompt))
	}

	return a.config.AuthCodeURL(state, params...), nil
}

// LoginHandler will redirect request to the login page.
func (a *Auth) LoginHandler(writer http.ResponseWriter, request *http.Request) {
	state := ""
	if a.StateCache != nil {
		var data *StateData
		var err error
		data, err = a.StateCache.CreateStateFromRequest(request)
		if err != nil {
			_ = httputil.WriteObj(writer, ec.NewError(ErrCodeStateGenerateFail, ErrMessageStateGenerateFail, err), http.StatusInternalServerError)
			return
		}

		state, err = a.StateCache.SaveState(data)
		if err != nil {
			_ = httputil.WriteObj(writer, ec.NewError(ErrCodeStateGenerateFail, ErrMessageStateGenerateFail, err), http.StatusInternalServerError)
			return
		}
	}

	q := request.URL.Query()
	prompt := q.Get("prompt")
	u, err := a.CreateAuthUrl(prompt, state)
	if err != nil {
		_ = httputil.WriteObj(writer, err, http.StatusInternalServerError)
		return
	}

	http.Redirect(writer, request, u, http.StatusSeeOther)
}

// OidcHandler handles request to the redirect URL registered in the identity provider.
func (a *Auth) OidcHandler(writer http.ResponseWriter, request *http.Request) {
	oidcError := request.FormValue("error")
	if oidcError != "" {
		if oidcError == "login_required" {
			_ = httputil.WriteObj(writer, ec.NewErrorBasic(ErrCodeLoginRequired, ErrMessageLoginRequired), http.StatusUnauthorized)
			return
		} else {
			_ = httputil.WriteObj(writer, ec.NewError(ErrCodeGenericOAuth2, ErrMessageGenericOAuth2, errors.New(oidcError)), http.StatusUnauthorized)
			return
		}
	}

	code := request.FormValue("code")
	if code == "" {
		_ = httputil.WriteObj(writer, ec.NewErrorBasic(ErrCodeNoOAuth2ExchangeCode, ErrMessageNoOAuth2ExchangeCode), http.StatusBadRequest)
		return
	}

	stateKey := request.FormValue("state")

	token, err := a.config.Exchange(context.Background(), code)
	if err != nil {
		a.Logger.Warnf("cannot exchange code for token: %s", strings.ReplaceAll(err.Error(), "\n", " "))
		_ = httputil.WriteObj(writer, ec.NewErrorBasic(ErrCodeOAuth2ExchangeFailed, ErrMessageOAuth2ExchangeFailed), http.StatusInternalServerError)
		return
	}

	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok {
		a.Logger.Warnf("cannot exchange code for token: %s", errors.New("token is missing from exchange").Error())
		_ = httputil.WriteObj(writer, ec.NewErrorBasic(ErrCodeOAuth2ExchangeFailed, ErrMessageOAuth2ExchangeFailed), http.StatusInternalServerError)
		return
	}

	idToken, _, err := a.VerifyRawIdToken(rawIdToken)
	if err != nil {
		a.Logger.Warnf("cannot verify ID token: %s", err)
		_ = httputil.WriteObj(writer, ec.NewError(ErrCodeCannotVerifyIdToken, ErrMessageCannotVerifyIdToken, err), http.StatusInternalServerError)
		return
	}

	if a.NonceCache != nil {
		err = a.NonceCache.VerifyNonce(idToken.Nonce)
		if err != nil && err != ErrNonceInvalid {
			_ = httputil.WriteObj(writer, ec.NewError(ErrCodeCannotVerifyIdToken, ErrMessageCannotVerifyIdToken, err), http.StatusInternalServerError)
			return
		}

		if err == ErrNonceInvalid {
			_ = httputil.WriteObj(writer, ec.NewError(ErrCodeGenericOAuth2, ErrMessageGenericOAuth2, err), http.StatusUnauthorized)
			return
		}
	}

	var stateData *StateData
	if a.StateCache != nil {
		stateData = new(StateData)
		err = a.StateCache.VerifyState(stateKey, stateData)
		if err != nil && err != ErrStateInvalid {
			_ = httputil.WriteObj(writer, ec.NewError(ErrCodeCannotVerifyIdToken, ErrMessageCannotVerifyIdToken, err), http.StatusInternalServerError)
			return
		}

		if err == ErrStateInvalid {
			_ = httputil.WriteObj(writer, ec.NewError(ErrCodeGenericOAuth2, ErrMessageGenericOAuth2, err), http.StatusUnauthorized)
			return
		}
	}

	claims := &AccessToken{}
	t, err := jwt.ParseSigned(token.AccessToken)

	if err != nil {
		a.Logger.Warnf("access token cannot be parsed: %s", err)
		_ = httputil.WriteObj(writer, ec.NewError(ErrCodeOAuth2ExchangeFailed, ErrMessageOAuth2ExchangeFailed, err), http.StatusInternalServerError)
		return
	}
	err = t.UnsafeClaimsWithoutVerification(&claims)
	if err != nil {
		a.Logger.Warnf("access token cannot be verified: %s", err)
		_ = httputil.WriteObj(writer, ec.NewError(ErrCodeOAuth2ExchangeFailed, ErrMessageOAuth2ExchangeFailed, err), http.StatusInternalServerError)
		return
	}

	roles := claims.RealmAccess.Roles
	var filteredRole []string
	for _, item := range roles {
		if containsSubstring(item, "manajemen-jf") {
			filteredRole = append(filteredRole, item)
		}
	}
	// _ = httputil.WriteObj200(writer, map[string]interface{}{
	// 	"role": roles,
	// })

	userId := strings.TrimPrefix(claims.PreferredUsername, "dummy:")
	asn, err := a.GetUserDetail(context.Background(), userId, "")
	if err != nil {
		a.Logger.Warnf("cannot get ASN detail: %s", err)
		_ = httputil.WriteObj(writer, err, http.StatusInternalServerError)
		return
	}

	if asn == nil {
		_ = httputil.WriteObj(writer, ec.NewErrorBasic(ErrCodeNotAsn, ErrMessageNotAsn), http.StatusForbidden)
		return
	}

	asn.AccessToken = claims
	asn.Username = userId

	err = a.LoginChecker.CheckLogin(writer, idToken, asn)
	if err != nil {
		return
	}

	if a.AccessTokenCache != nil {
		err = a.AccessTokenCache.SaveAccessToken(rawIdToken, claims)
		if err != nil {
			a.Logger.Warnf("cannot cache access token: %s", err)
			_ = httputil.WriteObj(writer, ec.NewErrorBasic(ErrCodeOAuth2ExchangeFailed, ErrMessageOAuth2ExchangeFailed), http.StatusInternalServerError)
			return
		}
	}

	http.SetCookie(writer, &http.Cookie{
		Name:     "token",
		Value:    rawIdToken,
		Expires:  token.Expiry,
		HttpOnly: a.CookieHttpOnly,
		SameSite: a.CookieSameSite,
		Secure:   a.CookieSecure,
		Domain:   a.CookieDomain,
	})

	userRole := strings.Join(filteredRole, ", ")

	// _ = httputil.WriteObj200(writer, map[string]interface{}{
	// 	"role": userRole,
	// })

	http.SetCookie(writer, &http.Cookie{
		Name:     "role",
		Value:    userRole,
		Expires:  token.Expiry,
		HttpOnly: a.CookieHttpOnly,
		SameSite: a.CookieSameSite,
		Secure:   a.CookieSecure,
		Domain:   a.CookieDomain,
	})

	redirectUrl := a.SuccessRedirectUrl
	if stateData != nil && stateData.ReturnUrl != "" {
		redirectUrl = stateData.ReturnUrl
	}

	http.Redirect(writer, request, redirectUrl, http.StatusSeeOther)
}

func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}

// LogoutHandler unconditionally removes the token cookie from the client, it will then redirect the user to the logout page.
// Redirect will be done with 303.
func (a *Auth) LogoutHandler(writer http.ResponseWriter, request *http.Request) {
	var token *http.Cookie
	var err error
	if token, err = request.Cookie("token"); err == http.ErrNoCookie {
		http.Redirect(writer, request, a.SuccessRedirectUrl, http.StatusSeeOther)
		return
	}

	http.SetCookie(writer, &http.Cookie{
		Name:     "token",
		MaxAge:   -1,
		HttpOnly: a.CookieHttpOnly,
		SameSite: a.CookieSameSite,
		Secure:   a.CookieSecure,
		Domain:   a.CookieDomain,
	})

	http.SetCookie(writer, &http.Cookie{
		Name:     "role",
		MaxAge:   -1,
		HttpOnly: a.CookieHttpOnly,
		SameSite: a.CookieSameSite,
		Secure:   a.CookieSecure,
		Domain:   a.CookieDomain,
	})

	if token == nil || token.Value == "" {
		http.Redirect(writer, request, a.SuccessRedirectUrl, http.StatusSeeOther)
		return
	}

	logoutUrl, _ := url.Parse(a.EndSessionEndpoint)
	q := url.Values{}
	q.Set("id_token_hint", token.Value)
	q.Set("state", uuid.New().String())
	//q.Set("post_logout_redirect_uri", a.SuccessRedirectUrl)
	logoutUrl.RawQuery = q.Encode()
	http.Redirect(writer, request, logoutUrl.String(), http.StatusSeeOther)
}

// WriteUnauthorized writes 401 with error message and removes the token cookie.
func (a *Auth) WriteUnauthorized(writer http.ResponseWriter, code int, errMessage string, err error) {
	http.SetCookie(writer, &http.Cookie{
		Name:     "token",
		MaxAge:   -1,
		HttpOnly: a.CookieHttpOnly,
		SameSite: a.CookieSameSite,
		Secure:   a.CookieSecure,
		Domain:   a.CookieDomain,
	})
	_ = httputil.WriteObj(writer, ec.NewError(code, errMessage, err), http.StatusUnauthorized)
}

// VerifyRawIdToken verifies raw ID token string and will return ID token object and user data extracted from the claims.
func (a *Auth) VerifyRawIdToken(rawIdToken string) (idToken *oidc.IDToken, user *User, err error) {
	idToken, err = a.verifier.Verify(context.Background(), rawIdToken)
	if err != nil {
		return
	}

	usr := &User{}
	err = idToken.Claims(&usr)
	if err != nil {
		return
	}

	return idToken, usr, nil
}

// VerifyUserinfo reads the token present in request cookie and verify the token.
// The user data is returned from the token.
func (a *Auth) VerifyUserinfo(writer http.ResponseWriter, request *http.Request) (user *User, err error) {
	rawIdTokenCookie, err := request.Cookie("token")
	if err != nil || rawIdTokenCookie == nil || rawIdTokenCookie.Value == "" {
		a.WriteUnauthorized(writer, ErrCodeNoIdToken, ErrMessageNoIdToken, err)
		return
	}

	_, user, err = a.VerifyRawIdToken(rawIdTokenCookie.Value)
	if err != nil {
		a.Logger.Warnf("cannot verify ID token: %s", err)
		a.WriteUnauthorized(writer, ErrCodeNoIdToken, ErrMessageNoIdToken, err)
		return
	}

	return user, nil
}

// UserAuthHandler is a middleware to verify user token cookie.
// It must carry an ID token and must be valid. The retrieved User information from ID token is then injected into
// the request object, which you can pull back using ReqGetUser function in your handler.
func (a *Auth) UserAuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		user, err := a.VerifyUserinfo(writer, request)
		if err != nil {
			return
		}

		next.ServeHTTP(writer, reqSetUser(request, user))
	})
}

// UserExtendedAuthHandler is a middleware to verify user token cookie, and to retrieve user extended data from cache.
// It must carry an ID token and must be valid. The retrieved User information from ID token is then injected into
// the request object, which you can pull back using ReqGetUser function in your handler.
//
// User will be given 401 if access token cannot be retrieved in the cache, you can treat this as session expired and
// redirect the user to login again.
//
// Cannot be used if AccessTokenCache is nil.
func (a *Auth) UserExtendedAuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		rawIdTokenCookie, err := request.Cookie("token")
		if err != nil || rawIdTokenCookie == nil || rawIdTokenCookie.Value == "" {
			a.WriteUnauthorized(writer, ErrCodeNoIdToken, ErrMessageNoIdToken, err)
			return
		}

		_, user, err := a.VerifyRawIdToken(rawIdTokenCookie.Value)
		if err != nil {
			a.Logger.Warnf("cannot verify ID token: %s", err)
			a.WriteUnauthorized(writer, ErrCodeNoIdToken, ErrMessageNoIdToken, err)
			return
		}

		accessToken, err := a.AccessTokenCache.GetAccessToken(rawIdTokenCookie.Value)
		if err != nil {
			a.Logger.Warnf("cannot retrieve user access token: %s", err)
			_ = httputil.WriteObj(writer, ec.NewError(ErrCodeCannotVerifyIdToken, ErrMessageCannotVerifyIdToken, err), http.StatusInternalServerError)
			return
		}
		if accessToken == nil {
			a.WriteUnauthorized(writer, ErrCodeNoUserInSessionStorage, ErrMessageNoUserInSessionStorage, err)
			return
		}
		user.AccessToken = accessToken

		next.ServeHTTP(writer, reqSetUser(request, user))
	})
}

// reqSetUserDetail adds user data to request.
// Given a request, attach user data to it via the request context.
func reqSetUser(request *http.Request, user *User) (newRequest *http.Request) {
	ctx := request.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, userContextKey, user)
	return request.WithContext(ctx)
}

// ReqGetUser retrieves User data that was injected to the request by UserAuthHandler.
// If no user was found, it will return nil, which you can interpret as HTTP status code 401. Although because you
// must be using UserAuthHandler as there is no other way to inject User information to the request, it has been
// handled in the handler.
func ReqGetUser(request *http.Request) (user *User) {
	ctx := request.Context()
	if ctx == nil {
		return nil
	}

	value := ctx.Value(userContextKey)
	if value == nil {
		return nil
	}

	if usr, ok := value.(*User); ok {
		return usr
	}

	return nil
}

// AssertReqGetUser retrieves user data but will panic if user is not found in the request, meaning
// that UserAuthHandler was not used.
func AssertReqGetUser(request *http.Request) (user *User) {
	user = ReqGetUser(request)
	if user == nil {
		panic("user is not available in request")
	}
	return user
}

// AssertReqGetUserExtended retrieves user data but will panic if user and access token is not found in the request, meaning
// that UserExtendedAuthHandler was not used.
func AssertReqGetUserExtended(request *http.Request) (user *User) {
	user = ReqGetUser(request)
	if user == nil {
		panic("user is not available in request")
	}

	if user.AccessToken == nil {
		panic("access token is not available in request")
	}
	return user
}

// UserDetailAuthHandler returns a middleware handler to verifies the request and return
// user extended information. It is the extension of UserExtendedAuthHandler.
//
// Cannot be used if AccessTokenCache is nil.
//
// User basic data are then used to lookup user extended details in the BKN database. The retrieved user
// detail is then attached back to the request for the next handler to use, which can be retrieved back with ReqGetUserDetail.
//
// Because it requires access to the database, it is advisable not to use it in every request.
//
// 403 and ErrNoWorkAgencyId will be returned also if the user does not have WorkAgencyId.
func (a *Auth) UserDetailAuthHandler(next http.Handler) http.Handler {
	return a.UserExtendedAuthHandler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		user := AssertReqGetUserExtended(request)

		userDetail, err := a.GetUserDetail(context.Background(), user.PreferredUsername, "")
		if err != nil {
			_ = httputil.WriteObj(writer, err, http.StatusInternalServerError)
			return
		}
		userDetail.Email = user.Email
		userDetail.Username = user.PreferredUsername
		userDetail.AccessToken = user.AccessToken

		if userDetail.WorkAgencyId == "" {
			_ = httputil.WriteObj(writer, ec.NewErrorBasic(ErrCodeNoWorkAgencyId, ErrMessageNoWorkAgencyId), http.StatusForbidden)
			return
		}

		if user == nil {
			_ = httputil.WriteObj(writer, ec.NewErrorBasic(ErrCodeInvalidUser, ErrMessageInvalidUser), http.StatusForbidden)
			return
		}

		next.ServeHTTP(writer, reqSetUserDetail(request, userDetail))
	}))
}

// reqSetUserDetail adds user detail data to request.
// Given a request, attach user detail data to it via the request context.
func reqSetUserDetail(request *http.Request, user *Asn) (newRequest *http.Request) {
	ctx := request.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, userDetailContextKey, user)
	return request.WithContext(ctx)
}

// ReqGetUserDetail retrieves user data previously attached by reqSetUserDetail from a request.
func ReqGetUserDetail(request *http.Request) (user *Asn) {
	ctx := request.Context()
	if ctx == nil {
		return nil
	}

	value := ctx.Value(userDetailContextKey)
	if value == nil {
		return nil
	}

	if usr, ok := value.(*Asn); ok {
		return usr
	}

	return nil
}

// AssertReqGetUserDetail retrieves user detail but will panic if user detail is not found in the request, meaning
// that UserDetailAuthHandler was not used.
func AssertReqGetUserDetail(request *http.Request) (user *Asn) {
	user = ReqGetUserDetail(request)
	if user == nil {
		panic("user detail is not available in request")
	}
	if user.AsnId == "" {
		panic("user detail in request is incomplete, no AsnId")
	}
	if user.WorkAgencyId == "" {
		panic("user detail in request is incomplete, no WorkAgencyId")
	}
	return user
}

// InjectUserExtended injects user extended data to the request.
// Can be used for unit testing.
func InjectUserExtended(request *http.Request, user *User) (newRequest *http.Request) {
	return reqSetUser(request, user)
}

// InjectUserDetail injects user detailed data to the request.
// Can be used for unit testing.
func InjectUserDetail(request *http.Request, asn *Asn) (newRequest *http.Request) {
	return reqSetUserDetail(request, asn)
}
