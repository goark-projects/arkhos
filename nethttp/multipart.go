package nethttp

import (
	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/multipart"
)

// MultipartParserFactory 创建 multipart 解析器。
type MultipartParserFactory func() *multipart.Parser

// NewMultipartParser 创建标准 multipart 解析器。
func NewMultipartParser(options ...multipart.Option) *multipart.Parser {
	return multipart.NewParser(options...)
}

// ParseMultipart 使用当前容器配置解析并绑定 multipart 表单。
func ParseMultipart(req *servlet.Request) (*multipart.Form, error) {
	profile, ok := currentProfile(req)
	if !ok || profile.multipartParser == nil {
		return nil, ErrMultipartProfileUnavailable
	}
	return multipart.ParseRequest(req, profile.multipartParser)
}

// MultipartForm 返回当前请求已解析的 multipart 表单。
func MultipartForm(req *servlet.Request) (*multipart.Form, bool) {
	return multipart.Current(req)
}

// MultipartPart 返回当前请求指定字段的第一个文件段。
func MultipartPart(req *servlet.Request, name string) (multipart.Part, bool, error) {
	profile, ok := currentProfile(req)
	if !ok || profile.multipartParser == nil {
		return multipart.Part{}, false, ErrMultipartProfileUnavailable
	}
	return multipart.RequestPart(req, name, profile.multipartParser)
}

// MultipartParts 返回当前请求的所有文件段。
func MultipartParts(req *servlet.Request) ([]multipart.Part, error) {
	profile, ok := currentProfile(req)
	if !ok || profile.multipartParser == nil {
		return nil, ErrMultipartProfileUnavailable
	}
	return multipart.Parts(req, profile.multipartParser)
}
