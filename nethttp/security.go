package nethttp

import (
	"context"
	"errors"
	"net/http"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/security"
)

// SecurityPolicy 在请求进入应用处理器前执行认证和授权。
type SecurityPolicy interface {
	Apply(ctx context.Context, req *servlet.Request, res servlet.Response) error
}

// SecurityPolicyFunc 将函数适配为安全策略。
type SecurityPolicyFunc func(ctx context.Context, req *servlet.Request, res servlet.Response) error

// Apply 执行安全策略。
func (f SecurityPolicyFunc) Apply(ctx context.Context, req *servlet.Request, res servlet.Response) error {
	return f(ctx, req, res)
}

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

// ConstraintSecurityPolicy 创建只执行声明式约束的安全策略。
func ConstraintSecurityPolicy(constraint security.Constraint) SecurityPolicy {
	return SecurityPolicyFunc(func(ctx context.Context, req *servlet.Request, _ servlet.Response) error {
		return constraint.Authorize(ctx, req)
	})
}
