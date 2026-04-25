package transformer

// Format 标识请求/响应的协议格式，用于注册表查找和透明代理判断。
type Format string

const (
	FormatOpenAIChat      Format = "openai"
	FormatOpenAIResponse  Format = "openai-response"
	FormatOpenAIEmbedding Format = "openai-embedding"
	FormatAnthropic       Format = "anthropic"
	FormatGemini          Format = "gemini"
	FormatVolcengine      Format = "volcengine"
)

// BaseFormat 返回格式的基础协议族。
// 用于透明代理判断：只要基础协议族相同，就可以透传。
// 例如 "openai" 和 "openai-response" 都属于不同的基础协议族，需要转换。
func (f Format) BaseFormat() Format {
	return f
}

// String 返回格式的字符串表示。
func (f Format) String() string {
	return string(f)
}
