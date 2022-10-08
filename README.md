## 一个简易消息传递服务

原始目标是 webrtc 的信令服务器, 但后面想了想, 其实可以做成简易的消息传递服务,
于是提取出来了作为单独的一个库/标准.

是 <https://github.com/rabbitmq/amqp091-go> 的简化版/弱化版

## 架构

- 任务处理方
  - 通过 `EventSource` 监听进来的请求并处理
  - 任务处理完成后通过 `http.MethodDelete` 返回处理结果
- 调用方
  - 通过 `http.MethodPost` 发送二进制消息到对应主题

注: 可以对二进制消息进行加密防止服务器对消息进行窥探

## 使用

可以参考 [signaler_test.go](./signaler_test.go) 文件中的 `TestSignaler` 函数

### 启动信令服务器

```go
package main

import (
	"net/http"

	"github.com/shynome/lens"
)

func main() {
	s := signaler.New(nil)
  http.Handle("/signaler", s)
	http.ListenAndServe(":7070", nil)
}
```

### 监听特定主题的消息并处理

```go
package main

import (
	"github.com/donovanhide/eventsource"
	"github.com/lainio/err2/try"
	"github.com/shynome/lens/sdk"
)

func main() {
  /*
  a:b 必不可少, 每个用户都有一个属于自己的信令服务
  t=7 必不可少, 监听/呼叫/处理对应主题的消息
  */
	var endpoint = "http://a:b@127.0.0.1:7070/signaler?t=7"
	sdk := sdk.New(endpoint)
	stream := try.To1(
		eventsource.Subscribe(endpoint, ""))
	for ev := range stream.Events {
    // 直接返回输入
		go sdk.Dial(ev.Id(), []byte(ev.Data()))
	}
}
```

### 使用

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/lainio/err2/try"
	"github.com/shynome/lens/sdk"
)

func main() {
	sdk := sdk.New("http://a:b@127.0.0.1:7070/signaler?t=7")
	var input = []byte("hello")
	rbody := try.To1(
		sdk.Call(input))
	if !bytes.Equal(rbody, input) {
		try.To(fmt.Errorf("expect %s, got %s", input, rbody))
	}
}
```
