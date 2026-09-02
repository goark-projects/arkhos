package hertz

import (
	internalprofile "goark.dev/arkhos/internal/profile"

	"goark.dev/arkarta/servlet/security"
)

// SecurityPolicy 在请求进入应用处理器前执行认证和授权。
type SecurityPolicy = internalprofile.SecurityPolicy

// SecurityPolicyFunc 将函数适配为安全策略。
type SecurityPolicyFunc = internalprofile.SecurityPolicyFunc

// BasicSecurityPolicy 创建 Basic 认证加声明式约束策略。
func BasicSecurityPolicy(authenticator security.Authenticator, constraint security.Constraint) SecurityPolicy {
	return internalprofile.BasicSecurityPolicy(authenticator, constraint)
}

// ConstraintSecurityPolicy 创建只执行声明式约束的安全策略。
func ConstraintSecurityPolicy(constraint security.Constraint) SecurityPolicy {
	return internalprofile.ConstraintSecurityPolicy(constraint)
}
