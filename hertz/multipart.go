package hertz

import (
	internalprofile "goark.dev/arkhos/internal/profile"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/multipart"
)

// MultipartParserFactory 创建 multipart 解析器。
type MultipartParserFactory = internalprofile.MultipartParserFactory

// NewMultipartParser 创建 multipart 解析器。
func NewMultipartParser(options ...multipart.Option) *multipart.Parser {
	return multipart.NewParser(options...)
}

// ParseMultipart 解析并缓存当前请求的 multipart 表单。
func ParseMultipart(req *servlet.Request) (*multipart.Form, error) {
	return internalprofile.ParseMultipart(req)
}

// MultipartForm 返回当前请求已经解析的 multipart 表单。
func MultipartForm(req *servlet.Request) (*multipart.Form, bool) {
	return multipart.Current(req)
}

// MultipartPart 返回指定名称的第一个 multipart 部分。
func MultipartPart(req *servlet.Request, name string) (multipart.Part, bool, error) {
	return internalprofile.MultipartPart(req, name)
}

// MultipartParts 返回当前请求的全部 multipart 部分。
func MultipartParts(req *servlet.Request) ([]multipart.Part, error) {
	return internalprofile.MultipartParts(req)
}
