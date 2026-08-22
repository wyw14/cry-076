# Bug Reproduction

## Bug 是什么

请求携带已有 request ID 时，未授权响应的错误体丢失该 ID，调用方无法用原请求标识关联错误。

## 如何触发

向受保护接口发送未授权请求，同时携带一个确定的 request ID，再读取错误响应体中的 request ID。

## 错误信息

```text
error body request id="", want incoming id
```
