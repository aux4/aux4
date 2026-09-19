package jwt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// testKeyPair generates a throwaway RSA key pair and writes a JWKS file
// exposing its public half, for fully offline (no network) test fixtures.
func testKeyPair(t *testing.T) (*rsa.PrivateKey, string, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating test key: %v", err)
	}

	kid := "test-kid-1"
	jwksPath := filepath.Join(t.TempDir(), "jwks.json")

	set := keySet{
		Keys: []jwk{
			{
				Kty: "RSA",
				Kid: kid,
				N:   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
				E:   base64.RawURLEncoding.EncodeToString(bigIntBytes(key.PublicKey.E)),
				Alg: "RS256",
			},
		},
	}

	data, err := json.Marshal(set)
	if err != nil {
		t.Fatalf("marshaling jwks: %v", err)
	}
	if err := os.WriteFile(jwksPath, data, 0644); err != nil {
		t.Fatalf("writing jwks file: %v", err)
	}

	return key, kid, jwksPath
}

func bigIntBytes(e int) []byte {
	// Standard RSA public exponent (65537) encodes to 3 bytes; this helper
	// keeps the test fixture generic instead of hardcoding "AQAB".
	b := []byte{byte(e >> 16), byte(e >> 8), byte(e)}
	// Trim leading zero bytes, matching how JWKS encodes E.
	for len(b) > 1 && b[0] == 0 {
		b = b[1:]
	}
	return b
}

func signToken(t *testing.T, key *rsa.PrivateKey, kid string, alg string, claims map[string]interface{}) string {
	t.Helper()

	header := map[string]interface{}{"alg": alg, "typ": "JWT", "kid": kid}
	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := headerB64 + "." + claimsB64

	if alg != "RS256" {
		// Used only to build a malformed/unsigned-style token for the alg
		// rejection test; the "signature" is meaningless.
		return signingInput + ".bm90LWEtc2lnbmF0dXJl"
	}

	hashed := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])
	if err != nil {
		t.Fatalf("signing test token: %v", err)
	}

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func baseClaims() map[string]interface{} {
	now := time.Now()
	return map[string]interface{}{
		"sub":   "user-123",
		"email": "user@example.com",
		"scope": "cloud:invoke",
		"iss":   "https://sso.aux4.io",
		"aud":   "aux4-machine-invoke",
		"iat":   now.Unix(),
		"exp":   now.Add(15 * time.Minute).Unix(),
	}
}

func TestVerifyRS256_ValidToken(t *testing.T) {
	key, kid, jwksPath := testKeyPair(t)
	token := signToken(t, key, kid, "RS256", baseClaims())

	claims, err := VerifyRS256(token, jwksPath, VerifyOptions{
		Issuer:    "https://sso.aux4.io",
		Scope:     "cloud:invoke",
		ClockSkew: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected valid token to verify, got error: %v", err)
	}
	if claims.String("sub") != "user-123" {
		t.Errorf("expected sub=user-123, got %q", claims.String("sub"))
	}
	if claims.String("email") != "user@example.com" {
		t.Errorf("expected email claim to survive, got %q", claims.String("email"))
	}
}

func TestVerifyRS256_EmptyToken(t *testing.T) {
	_, _, jwksPath := testKeyPair(t)
	if _, err := VerifyRS256("", jwksPath, VerifyOptions{}); err == nil {
		t.Fatal("expected empty token to fail closed")
	}
}

func TestVerifyRS256_MalformedToken(t *testing.T) {
	_, _, jwksPath := testKeyPair(t)
	if _, err := VerifyRS256("not-a-jwt", jwksPath, VerifyOptions{}); err == nil {
		t.Fatal("expected malformed token to fail closed")
	}
}

func TestVerifyRS256_ExpiredToken(t *testing.T) {
	key, kid, jwksPath := testKeyPair(t)
	claims := baseClaims()
	claims["exp"] = time.Now().Add(-1 * time.Hour).Unix()
	token := signToken(t, key, kid, "RS256", claims)

	if _, err := VerifyRS256(token, jwksPath, VerifyOptions{ClockSkew: 30 * time.Second}); err == nil {
		t.Fatal("expected expired token to fail closed")
	}
}

func TestVerifyRS256_MissingExpClaim(t *testing.T) {
	key, kid, jwksPath := testKeyPair(t)
	claims := baseClaims()
	delete(claims, "exp")
	token := signToken(t, key, kid, "RS256", claims)

	if _, err := VerifyRS256(token, jwksPath, VerifyOptions{}); err == nil {
		t.Fatal("expected token with no exp claim to fail closed")
	}
}

func TestVerifyRS256_WrongSignature(t *testing.T) {
	key, kid, jwksPath := testKeyPair(t)
	token := signToken(t, key, kid, "RS256", baseClaims())

	// Sign with a DIFFERENT key but publish the original key's JWKS — the
	// signature must not verify against a key that didn't produce it.
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating imposter key: %v", err)
	}
	imposterToken := signToken(t, otherKey, kid, "RS256", baseClaims())
	_ = token

	if _, err := VerifyRS256(imposterToken, jwksPath, VerifyOptions{}); err == nil {
		t.Fatal("expected signature from a non-matching key to fail closed")
	}
}

func TestVerifyRS256_RejectsNoneAlg(t *testing.T) {
	key, kid, jwksPath := testKeyPair(t)
	token := signToken(t, key, kid, "none", baseClaims())

	if _, err := VerifyRS256(token, jwksPath, VerifyOptions{}); err == nil {
		t.Fatal("expected alg=none token to be rejected (fail closed)")
	}
}

func TestVerifyRS256_UnknownKid(t *testing.T) {
	key, _, jwksPath := testKeyPair(t)
	token := signToken(t, key, "some-other-kid", "RS256", baseClaims())

	if _, err := VerifyRS256(token, jwksPath, VerifyOptions{}); err == nil {
		t.Fatal("expected unknown kid to fail closed")
	}
}

func TestVerifyRS256_ScopeMismatch(t *testing.T) {
	key, kid, jwksPath := testKeyPair(t)
	token := signToken(t, key, kid, "RS256", baseClaims())

	if _, err := VerifyRS256(token, jwksPath, VerifyOptions{Scope: "cloud:build"}); err == nil {
		t.Fatal("expected scope mismatch to fail closed")
	}
}

func TestVerifyRS256_IssuerMismatch(t *testing.T) {
	key, kid, jwksPath := testKeyPair(t)
	token := signToken(t, key, kid, "RS256", baseClaims())

	if _, err := VerifyRS256(token, jwksPath, VerifyOptions{Issuer: "https://evil.example.com"}); err == nil {
		t.Fatal("expected issuer mismatch to fail closed")
	}
}

func TestVerifyRS256_MissingJwksFile(t *testing.T) {
	key, kid, _ := testKeyPair(t)
	token := signToken(t, key, kid, "RS256", baseClaims())

	if _, err := VerifyRS256(token, "/nonexistent/jwks.json", VerifyOptions{}); err == nil {
		t.Fatal("expected a missing jwks file to fail closed, not silently pass")
	}
}
