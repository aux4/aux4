// Package jwt verifies RS256-signed JWTs entirely locally against a JWKS
// (JSON Web Key Set) already present on disk. It performs NO network I/O —
// fetching and caching the JWKS is the caller's responsibility (the aux4
// standard library has no HTTP client dependency here on purpose, so this
// package stays a pure, easily-testable crypto/claims verifier built only on
// the Go standard library — no third-party JWT/JOSE dependency, no added
// image weight for anything that already ships the aux4 binary).
package jwt

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

// Claims is the decoded JWT payload (the second segment).
type Claims map[string]interface{}

// String returns the named claim as a string, or "" if it is absent or not
// a string.
func (c Claims) String(name string) string {
	if v, ok := c[name]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

type keySet struct {
	Keys []jwk `json:"keys"`
}

// VerifyOptions controls claim validation beyond the signature itself. Any
// field left at its zero value ("" for strings) skips that check.
type VerifyOptions struct {
	// Issuer, if non-empty, must exactly match the token's "iss" claim.
	Issuer string
	// Audience, if non-empty, must appear in the token's "aud" claim (which
	// may be a single string or an array of strings per the JWT spec).
	Audience string
	// Scope, if non-empty, must be one of the space-delimited tokens in the
	// token's "scope" claim (OAuth2 scope string, RFC 6749 §3.3).
	Scope string
	// ClockSkew is the leeway applied to exp/nbf checks.
	ClockSkew time.Duration
}

// VerifyRS256 verifies the RS256 signature of token against the RSA public
// keys in the JWKS file at jwksFilePath (matched by the token's header "kid"),
// then validates the standard time claims (exp required, nbf if present) and
// whatever is requested in opts. It returns the decoded claims only when
// every check passes; any failure — malformed token, unsupported/absent alg,
// no matching key, bad signature, expired, or an option mismatch — returns a
// non-nil error and nil claims. There is no partial-success case: callers
// must treat any error as "unauthenticated", never fall through to an
// unscoped identity.
func VerifyRS256(token string, jwksFilePath string, opts VerifyOptions) (Claims, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("empty token")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed token: expected 3 dot-separated segments")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("malformed header: %w", err)
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("malformed payload: %w", err)
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("malformed signature: %w", err)
	}

	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("malformed header json: %w", err)
	}

	// Only RS256 is accepted. This is deliberate, not an oversight: accepting
	// "none" would skip signing entirely, and accepting an HMAC alg (HS256)
	// would let a caller forge a token using the RSA PUBLIC key published in
	// the JWKS as if it were an HMAC secret — the classic alg-confusion
	// attack against JWKS-based verifiers.
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported alg %q (only RS256 is accepted)", header.Alg)
	}

	pub, err := findRSAKey(jwksFilePath, header.Kid)
	if err != nil {
		return nil, err
	}

	signingInput := parts[0] + "." + parts[1]
	hashed := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], sig); err != nil {
		return nil, errors.New("invalid signature")
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("malformed claims json: %w", err)
	}

	if err := validateClaims(claims, opts); err != nil {
		return nil, err
	}

	return claims, nil
}

func findRSAKey(jwksFilePath, kid string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(jwksFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading jwks file: %w", err)
	}

	var set keySet
	if err := json.Unmarshal(data, &set); err != nil {
		return nil, fmt.Errorf("malformed jwks file: %w", err)
	}

	for _, k := range set.Keys {
		if k.Kty != "RSA" {
			continue
		}
		if kid != "" && k.Kid != kid {
			continue
		}
		return buildRSAPublicKey(k)
	}

	return nil, fmt.Errorf("no matching RSA key found in jwks for kid %q", kid)
}

func buildRSAPublicKey(k jwk) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("malformed jwks modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("malformed jwks exponent: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
}

func validateClaims(claims Claims, opts VerifyOptions) error {
	skew := opts.ClockSkew
	now := time.Now()

	exp, ok := numericClaim(claims, "exp")
	if !ok {
		return errors.New("token has no exp claim")
	}
	if now.After(time.Unix(exp, 0).Add(skew)) {
		return errors.New("token expired")
	}

	if nbf, ok := numericClaim(claims, "nbf"); ok {
		if now.Before(time.Unix(nbf, 0).Add(-skew)) {
			return errors.New("token not yet valid")
		}
	}

	if opts.Issuer != "" && claims.String("iss") != opts.Issuer {
		return fmt.Errorf("unexpected issuer %q", claims.String("iss"))
	}

	if opts.Audience != "" && !audienceMatches(claims["aud"], opts.Audience) {
		return errors.New("unexpected audience")
	}

	if opts.Scope != "" && !scopeContains(claims.String("scope"), opts.Scope) {
		return fmt.Errorf("token scope does not include %q", opts.Scope)
	}

	return nil
}

func numericClaim(claims Claims, name string) (int64, bool) {
	v, ok := claims[name]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	case string:
		i, err := strconv.ParseInt(n, 10, 64)
		return i, err == nil
	}
	return 0, false
}

func audienceMatches(aud interface{}, expected string) bool {
	switch v := aud.(type) {
	case string:
		return v == expected
	case []interface{}:
		for _, a := range v {
			if s, ok := a.(string); ok && s == expected {
				return true
			}
		}
	}
	return false
}

// scopeContains checks whether expected is one of the whitespace-delimited
// tokens in scopeClaim (OAuth2 space-delimited scope string, RFC 6749 §3.3).
func scopeContains(scopeClaim, expected string) bool {
	for _, s := range strings.Fields(scopeClaim) {
		if s == expected {
			return true
		}
	}
	return false
}
