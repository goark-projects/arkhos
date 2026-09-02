package profile

import (
	"context"
	"errors"
	"net/http"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/security"
)

// BasicSecurityPolicy 创建 Basic 认证加声明式约束策略。
func BasicSecurityPolicy(authenticator security.Authenticator, constraint security.Constraint) SecurityPolicy {
	return SecurityPolicyFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
		ok, err := security.Authenticate(ctx, req, res, authenticator)
		if err != nil {
			if errors.Is(err, security.ErrAuthenticationFailed) {
				return servlet.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
			}
			return err
		}
		if !ok {
			return servlet.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), nil)
		}
		return constraint.Authorize(ctx, req)
	})
}

// ConstraintSecurityPolicy 创建只执行声明式约束的策略。
func ConstraintSecurityPolicy(constraint security.Constraint) SecurityPolicy {
	return SecurityPolicyFunc(func(ctx context.Context, req *servlet.Request, _ servlet.Response) error {
		return constraint.Authorize(ctx, req)
	})
}
