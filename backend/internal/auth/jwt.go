package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// JWTValidator validates JWT tokens from OAuth2 providers
type JWTValidator struct {
	issuerURL  string // Expected issuer claim in JWT
	jwksURL    string // URL to fetch JWKS (may differ from issuerURL for cluster-internal access)
	audience   string
	logger     *zap.Logger
	httpClient *http.Client
	jwksCache  map[string]*rsa.PublicKey
	cacheTTL   time.Time
	mutex      sync.RWMutex
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// NewJWTValidator creates a new JWT validator
// issuerURL: expected issuer claim in JWT tokens
// jwksURL: URL to fetch JWKS (if empty, derived from issuerURL)
// audience: expected audience claim (optional)
func NewJWTValidator(issuerURL, jwksURL, audience string, logger *zap.Logger) *JWTValidator {
	// If jwksURL not provided, derive from issuerURL
	if jwksURL == "" {
		jwksURL = fmt.Sprintf("%s/.well-known/jwks.json", issuerURL)
	}

	logger.Debug("initializing JWT validator",
		zap.String("issuer", issuerURL),
		zap.String("jwks_url", jwksURL),
		zap.String("audience", audience))

	return &JWTValidator{
		issuerURL:  issuerURL,
		jwksURL:    jwksURL,
		audience:   audience,
		logger:     logger,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		jwksCache:  make(map[string]*rsa.PublicKey),
	}
}

// ValidateToken validates a JWT token and returns the parsed token if valid
func (v *JWTValidator) ValidateToken(tokenString string) (*jwt.Token, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token is empty")
	}

	// Parse token without validation first to get kid
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		v.logger.Debug("failed to parse token header",
			zap.Error(err))
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Get kid from header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		v.logger.Debug("kid header missing in token")
		return nil, fmt.Errorf("kid header missing")
	}

	// Get public key for this kid
	publicKey, err := v.getPublicKey(kid)
	if err != nil {
		v.logger.Warn("failed to get public key",
			zap.String("kid", kid),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse and validate token with public key
	token, err = jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		v.logger.Debug("failed to validate token signature",
			zap.Error(err))
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	// Validate claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		v.logger.Debug("invalid token claims")
		return nil, fmt.Errorf("invalid token claims")
	}

	// Verify issuer
	iss, ok := claims["iss"].(string)
	if !ok || iss != v.issuerURL {
		v.logger.Debug("invalid issuer",
			zap.String("expected", v.issuerURL),
			zap.String("got", iss))
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", v.issuerURL, iss)
	}

	// Verify audience (if configured)
	if v.audience != "" {
		if !v.verifyAudience(claims) {
			v.logger.Debug("invalid audience",
				zap.String("expected", v.audience))
			return nil, fmt.Errorf("invalid audience")
		}
	}

	// Verify expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		v.logger.Debug("exp claim missing")
		return nil, fmt.Errorf("exp claim missing")
	}
	if time.Now().Unix() > int64(exp) {
		v.logger.Debug("token expired",
			zap.Time("expired_at", time.Unix(int64(exp), 0)))
		return nil, fmt.Errorf("token expired")
	}

	v.logger.Debug("token validated successfully",
		zap.String("issuer", iss),
		zap.String("subject", fmt.Sprintf("%v", claims["sub"])))

	return token, nil
}

// verifyAudience checks if the audience claim matches expected audience
func (v *JWTValidator) verifyAudience(claims jwt.MapClaims) bool {
	aud, ok := claims["aud"]
	if !ok {
		return false
	}

	// Audience can be string or array of strings
	switch audVal := aud.(type) {
	case string:
		return audVal == v.audience
	case []interface{}:
		for _, a := range audVal {
			if audStr, ok := a.(string); ok && audStr == v.audience {
				return true
			}
		}
	}

	return false
}

// getPublicKey retrieves the public key for the given kid
func (v *JWTValidator) getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check cache first (with TTL check)
	v.mutex.RLock()
	if time.Now().Before(v.cacheTTL) {
		if key, ok := v.jwksCache[kid]; ok {
			v.mutex.RUnlock()
			v.logger.Debug("using cached public key", zap.String("kid", kid))
			return key, nil
		}
	}
	v.mutex.RUnlock()

	// Fetch JWKS
	v.logger.Debug("fetching JWKS", zap.String("url", v.jwksURL))

	resp, err := v.httpClient.Get(v.jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS from %s: %w", v.jwksURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	v.logger.Debug("fetched JWKS", zap.Int("key_count", len(jwks.Keys)))

	// Find matching key and parse all keys into cache
	v.mutex.Lock()
	defer v.mutex.Unlock()

	// Clear old cache
	v.jwksCache = make(map[string]*rsa.PublicKey)

	var targetKey *rsa.PublicKey
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" {
			v.logger.Debug("skipping non-RSA key",
				zap.String("kid", key.Kid),
				zap.String("kty", key.Kty))
			continue
		}

		publicKey, err := v.parseRSAPublicKey(key.N, key.E)
		if err != nil {
			v.logger.Warn("failed to parse public key",
				zap.String("kid", key.Kid),
				zap.Error(err))
			continue
		}

		// Cache the key
		v.jwksCache[key.Kid] = publicKey

		if key.Kid == kid {
			targetKey = publicKey
		}
	}

	// Set cache TTL to 1 hour
	v.cacheTTL = time.Now().Add(1 * time.Hour)

	if targetKey == nil {
		return nil, fmt.Errorf("key with kid %s not found in JWKS", kid)
	}

	v.logger.Debug("cached public keys",
		zap.Int("count", len(v.jwksCache)),
		zap.String("target_kid", kid))

	return targetKey, nil
}

// parseRSAPublicKey parses base64url encoded n and e into RSA public key
func (v *JWTValidator) parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	// Decode n (modulus)
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode n: %w", err)
	}

	// Decode e (exponent)
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode e: %w", err)
	}

	// Create big.Int from bytes
	n := new(big.Int).SetBytes(nBytes)

	// Convert exponent bytes to int
	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 + int(b)
	}

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: eInt,
	}

	return publicKey, nil
}
