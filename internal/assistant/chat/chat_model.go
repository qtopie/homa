package chat

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ChatModel实现eino/components/model/interface.go接口
type HomaChatModel struct {
	client     *http.Client
	apiKey     string
	baseURL    string
	model      string
	timeout    time.Duration
	retryCount int
}

type HomaChatModelConfig struct {
	APIKey string
}

type MyChatModelOptions struct {
	Options    *model.Options
	RetryCount int
	Timeout    time.Duration
}

func NewHomaChatModel(config *HomaChatModelConfig) (*HomaChatModel, error) {
	if config.APIKey == "" {
		return nil, errors.New("api key is required")
	}

	return &HomaChatModel{
		client: &http.Client{},
		apiKey: config.APIKey,
	}, nil
}

func (m *HomaChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// 1. 处理选项
	options := &MyChatModelOptions{
		Options: &model.Options{
			Model: &m.model,
		},
		RetryCount: m.retryCount,
		Timeout:    m.timeout,
	}
	options.Options = model.GetCommonOptions(options.Options, opts...)
	options = model.GetImplSpecificOptions(options, opts...)

	// 2. 开始生成前的回调
	ctx = callbacks.OnStart(ctx, &model.CallbackInput{
		Messages: messages,
		Config: &model.Config{
			Model: *options.Options.Model,
		},
	})

	// 3. 执行生成逻辑
	response, err := m.doGenerate(ctx, messages, options)

	// 4. 处理错误和完成回调
	if err != nil {
		ctx = callbacks.OnError(ctx, err)
		return nil, err
	}

	ctx = callbacks.OnEnd(ctx, &model.CallbackOutput{
		Message: response,
	})

	return response, nil
}

func (m *HomaChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// 1. 处理选项
	options := &MyChatModelOptions{
		Options: &model.Options{
			Model: &m.model,
		},
		RetryCount: m.retryCount,
		Timeout:    m.timeout,
	}
	options.Options = model.GetCommonOptions(options.Options, opts...)
	options = model.GetImplSpecificOptions(options, opts...)

	// 2. 开始流式生成前的回调
	ctx = callbacks.OnStart(ctx, &model.CallbackInput{
		Messages: messages,
		Config: &model.Config{
			Model: *options.Options.Model,
		},
	})

	// 3. 创建流式响应
	// Pipe产生一个StreamReader和一个StreamWrite，向StreamWrite中写入可以从StreamReader中读到，二者并发安全。
	// 实现中异步向StreamWrite中写入生成内容，返回StreamReader作为返回值
	// ***StreamReader是一个数据流，仅可读一次，组件自行实现Callback时，既需要通过OnEndWithCallbackOutput向callback传递数据流，也需要向返回一个数据流，需要对数据流进行一次拷贝
	// 考虑到此种情形总是需要拷贝数据流，OnEndWithCallbackOutput函数会在内部拷贝并返回一个未被读取的流
	// 以下代码演示了一种流处理方式，处理方式不唯一
	sr, sw := schema.Pipe[*model.CallbackOutput](1)

	// 4. 启动异步生成
	go func() {
		defer sw.Close()

		// 流式写入
		m.doStream(ctx, messages, options, sw)
	}()

	// 5. 完成回调
	_, nsr := callbacks.OnEndWithStreamOutput(ctx, sr)

	return schema.StreamReaderWithConvert(nsr, func(t *model.CallbackOutput) (*schema.Message, error) {
		return t.Message, nil
	}), nil
}

func (m *HomaChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	// 实现工具绑定逻辑
	return nil, nil
}

func (m *HomaChatModel) doGenerate(ctx context.Context, messages []*schema.Message, opts *MyChatModelOptions) (*schema.Message, error) {
	// 实现生成逻辑
	return nil, nil
}

func (m *HomaChatModel) doStream(ctx context.Context, messages []*schema.Message, opts *MyChatModelOptions, sr *schema.StreamWriter[*model.CallbackOutput]) {
	// 流式生成文本写入sr中
	return
}
