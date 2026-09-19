package executor

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/jwt"
	"aux4.dev/aux4/output"
)

// Aux4JwtVerifyExecutor backs `aux4 aux4 jwt-verify`: local (no network)
// RS256 verification of a JWT against a JWKS file already on disk, used by
// the cloud command-runtime (CSEC-039) to authenticate the machine-invoke
// token without adding Node or a JOSE library to the image. Every failure
// path returns a non-nil error (core.AccessDeniedError) and prints nothing
// to stdout — callers must treat "no claims JSON on stdout" as unauthenticated,
// never fall through to an unscoped identity.
type Aux4JwtVerifyExecutor struct {
}

func (executor *Aux4JwtVerifyExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	token, err := jwtStringParam(params, command, actions, "token")
	if err != nil {
		return err
	}
	jwksFile, err := jwtStringParam(params, command, actions, "jwksFile")
	if err != nil {
		return err
	}
	issuer, err := jwtStringParam(params, command, actions, "issuer")
	if err != nil {
		return err
	}
	audience, err := jwtStringParam(params, command, actions, "audience")
	if err != nil {
		return err
	}
	scope, err := jwtStringParam(params, command, actions, "scope")
	if err != nil {
		return err
	}
	clockSkewRaw, err := jwtStringParam(params, command, actions, "clockSkew")
	if err != nil {
		return err
	}

	if token == "" {
		return core.AccessDeniedError(errors.New("missing token"))
	}
	if jwksFile == "" {
		return core.AccessDeniedError(errors.New("missing jwksFile"))
	}

	skewSeconds := 30
	if clockSkewRaw != "" {
		if v, err := strconv.Atoi(clockSkewRaw); err == nil {
			skewSeconds = v
		}
	}

	claims, err := jwt.VerifyRS256(token, jwksFile, jwt.VerifyOptions{
		Issuer:    issuer,
		Audience:  audience,
		Scope:     scope,
		ClockSkew: time.Duration(skewSeconds) * time.Second,
	})
	if err != nil {
		return core.AccessDeniedError(err)
	}

	out, err := json.Marshal(claims)
	if err != nil {
		return core.InternalError("failed to encode claims", err)
	}

	output.Out(output.StdOut).Println(string(out))
	return nil
}

// jwtStringParam resolves a declared variable through the normal aux4 lookup
// chain (arg > env > config > default), so this command behaves like every
// other aux4 command rather than only reading raw argv.
func jwtStringParam(params *param.Parameters, command core.Command, actions []string, name string) (string, error) {
	value, err := params.Get(command, actions, name)
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", nil
	}
	if s, ok := value.(string); ok {
		return s, nil
	}
	return "", nil
}
