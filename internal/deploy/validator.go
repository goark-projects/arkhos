package deploy

import (
	"errors"

	servletcontainer "goark.dev/arkarta/servlet/container"
)

// ErrUnsupportedProfile 表示部署声明了当前容器不支持的 Arkarta Profile。
var ErrUnsupportedProfile = errors.New("arkhos/internal/deploy: unsupported Arkarta profile")

// Validator 校验部署描述与容器能力是否匹配。
type Validator struct {
	metadata servletcontainer.Metadata
}

// NewValidator 创建基于容器元数据的部署校验器。
func NewValidator(metadata servletcontainer.Metadata) Validator {
	return Validator{metadata: metadata}
}

// Validate 校验部署描述是否可以被当前容器承载。
func (v Validator) Validate(deployment *servletcontainer.Deployment) error {
	if deployment == nil {
		return servletcontainer.ErrNilDeployment
	}
	for _, profile := range deployment.Profiles() {
		if !v.metadata.Supports(profile) {
			return ErrUnsupportedProfile
		}
	}
	return nil
}
