package auth

import (
	"context"
	"fmt"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"net/http"
	"reflect"
	"sync"
)

// AccessToken is a structure containing fields that are available in decoded access token.
// It is not generic, and works only with BKN Keycloak system.
type AccessToken struct {
	Acr            string   `json:"acr"`
	AllowedOrigins []string `json:"allowed-origins"`
	// Aud could be a string, but could also be an array.
	Aud               interface{} `json:"aud"`
	AuthTime          int64       `json:"auth_time"`
	Azp               string      `json:"azp"`
	Email             string      `json:"email"`
	EmailVerified     bool        `json:"email_verified"`
	Exp               int64       `json:"exp"`
	FamilyName        string      `json:"family_name"`
	GivenName         string      `json:"given_name"`
	Iat               int64       `json:"iat"`
	Iss               string      `json:"iss"`
	Jti               string      `json:"jti"`
	Name              string      `json:"name"`
	Nonce             string      `json:"nonce"`
	PreferredUsername string      `json:"preferred_username"`
	RealmAccess       *struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess map[string]interface{} `json:"resource_access"`
	Scope          string                 `json:"scope"`
	SessionState   string                 `json:"session_state"`
	Sub            string                 `json:"sub"`
	Typ            string                 `json:"typ"`
	Optional       interface{}            `json:"optional"`
}

type AccessTokenRole struct {
	Role     string `json:"role"`
	RoleType string `json:"role_type"`
}

func (token *AccessToken) GetRoles() (roles []*AccessTokenRole) {
	if token.RealmAccess != nil && token.RealmAccess.Roles != nil {
		for _, realmRole := range token.RealmAccess.Roles {
			roles = append(roles, &AccessTokenRole{
				Role:     realmRole,
				RoleType: "realm",
			})
		}
	}

	if token.ResourceAccess != nil {
		for key, value := range token.ResourceAccess {
			var v map[string]interface{}
			var ok bool
			if v, ok = value.(map[string]interface{}); !ok {
				continue
			}

			var r []interface{}
			if rawRoles, ok := v["roles"]; ok {
				if r, ok = rawRoles.([]interface{}); !ok {
					continue
				}
			} else {
				continue
			}

			for _, roleKey := range r {
				if k, ok := roleKey.(string); ok {
					roles = append(roles, &AccessTokenRole{
						Role:     k,
						RoleType: key,
					})
				}
			}
		}
	}

	return roles
}

type AccessTokenCache interface {
	// SaveAccessToken saves access token in cache.
	SaveAccessToken(idToken string, accessToken *AccessToken) (err error)

	// GetAccessToken retrieves access token from cache.
	GetAccessToken(idToken string) (accessToken *AccessToken, err error)

	// DeleteAccessToken deletes access token from cache.
	DeleteAccessToken(idToken string) (err error)
}

// MemoryAccessTokenCache implements AccessTokenCache by simply storing token data in a Golang map.
// Access to the map is protected with RWMutex which locks the whole map.
// MemoryAccessTokenCache also implements NonceCache.
type MemoryAccessTokenCache struct {
	mu        sync.RWMutex
	data      map[string]*AccessToken
	nonceData map[string]struct{}
	stateData map[string]*StateData
}

func NewMemoryAccessTokenCache() *MemoryAccessTokenCache {
	return &MemoryAccessTokenCache{
		data:      make(map[string]*AccessToken),
		nonceData: make(map[string]struct{}),
		stateData: make(map[string]*StateData),
	}
}

func (m *MemoryAccessTokenCache) CreateStateFromRequest(request *http.Request) (data *StateData, err error) {
	returnUrl := request.FormValue("return_url")
	return &StateData{ReturnUrl: returnUrl}, nil
}

func (m *MemoryAccessTokenCache) SaveState(data *StateData) (stateKey string, err error) {
	stateKey = uuid.NewString()
	m.stateData[stateKey] = data
	return stateKey, nil
}

func (m *MemoryAccessTokenCache) VerifyState(state string, data *StateData) (err error) {
	stateItem, ok := m.stateData[state]
	if !ok {
		return ErrStateInvalid
	}

	reflect.ValueOf(data).Elem().Set(reflect.ValueOf(stateItem))

	delete(m.stateData, state)

	return nil
}

func (m *MemoryAccessTokenCache) GenerateNonce() (nonce string, err error) {
	nonce = uuid.NewString()
	m.nonceData[nonce] = struct{}{}
	return nonce, nil
}

func (m *MemoryAccessTokenCache) VerifyNonce(nonce string) (err error) {
	_, ok := m.nonceData[nonce]
	if !ok {
		return ErrNonceInvalid
	}

	delete(m.nonceData, nonce)

	return nil
}

func (m *MemoryAccessTokenCache) SaveAccessToken(idToken string, accessToken *AccessToken) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[idToken] = accessToken
	return nil
}

func (m *MemoryAccessTokenCache) GetAccessToken(idToken string) (accessToken *AccessToken, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	token, ok := m.data[idToken]
	if !ok {
		return nil, nil
	}
	return token, nil
}

func (m *MemoryAccessTokenCache) DeleteAccessToken(idToken string) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, idToken)
	return nil
}

type RedisAccessTokenCache struct {
	Cache       *cache.Cache
	Prefix      string
	NoncePrefix string
	StatePrefix string
}

func NewRedisAccessTokenCache(cache *cache.Cache) *RedisAccessTokenCache {
	return &RedisAccessTokenCache{
		Cache:       cache,
		Prefix:      "access",
		NoncePrefix: "nonce",
		StatePrefix: "state",
	}
}

func (r *RedisAccessTokenCache) key(id string) string {
	return fmt.Sprintf("%s:%s", r.Prefix, id)
}

func (r *RedisAccessTokenCache) nonceKey(nonce string) string {
	return fmt.Sprintf("%s:%s", r.NoncePrefix, nonce)
}

func (r *RedisAccessTokenCache) stateKey(stateKey string) string {
	return fmt.Sprintf("%s:%s", r.StatePrefix, stateKey)
}

func (r *RedisAccessTokenCache) SaveState(data *StateData) (stateKey string, err error) {
	stateKey = uuid.NewString()
	err = r.Cache.Set(&cache.Item{
		Ctx:   context.Background(),
		Key:   r.stateKey(stateKey),
		Value: data,
	})
	if err != nil {
		return "", err
	}

	return stateKey, nil
}

func (r *RedisAccessTokenCache) CreateStateFromRequest(request *http.Request) (data *StateData, err error) {
	returnUrl := request.FormValue("return_url")
	return &StateData{ReturnUrl: returnUrl}, nil
}

func (r *RedisAccessTokenCache) VerifyState(stateKey string, data *StateData) (err error) {
	key := r.stateKey(stateKey)
	err = r.Cache.Get(context.Background(), key, data)
	if err != nil && err != redis.Nil && err != cache.ErrCacheMiss {
		return err
	}

	if err == redis.Nil || err == cache.ErrCacheMiss {
		return ErrStateInvalid
	}

	defer r.Cache.Delete(context.Background(), key)

	return nil
}

func (r *RedisAccessTokenCache) GenerateNonce() (nonce string, err error) {
	nonce = uuid.NewString()
	err = r.Cache.Set(&cache.Item{
		Ctx: context.Background(),
		Key: r.nonceKey(nonce),
	})
	if err != nil {
		return "", err
	}

	return nonce, err
}

func (r *RedisAccessTokenCache) VerifyNonce(nonce string) (err error) {
	key := r.nonceKey(nonce)
	err = r.Cache.Get(context.Background(), key, nil)
	if err != nil && err != redis.Nil && err != cache.ErrCacheMiss {
		return err
	}

	if err == redis.Nil || err == cache.ErrCacheMiss {
		return ErrNonceInvalid
	}

	defer r.Cache.Delete(context.Background(), key)

	return nil
}

func (r *RedisAccessTokenCache) SaveAccessToken(idToken string, accessToken *AccessToken) (err error) {
	return r.Cache.Set(&cache.Item{
		Ctx:   context.Background(),
		Key:   r.key(idToken),
		Value: accessToken,
	})
}

func (r *RedisAccessTokenCache) GetAccessToken(idToken string) (accessToken *AccessToken, err error) {
	accessToken = &AccessToken{}
	err = r.Cache.Get(context.Background(), r.key(idToken), accessToken)
	if err != nil {
		if err == redis.Nil || err == cache.ErrCacheMiss {
			return nil, nil
		}
	}

	return accessToken, nil
}

func (r *RedisAccessTokenCache) DeleteAccessToken(idToken string) (err error) {
	err = r.Cache.Delete(context.Background(), r.key(idToken))
	if err == redis.Nil || err == cache.ErrCacheMiss {
		err = nil
		return err
	}

	return nil
}
