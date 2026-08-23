# Bug Reproduction

## Bug 是什么

保存成员档案成功后，响应中的版本号仍是保存前的旧版本。客户端继续使用该版本提交下一次修改时会发生版本冲突。

## 如何触发

使用版本 7 保存一次档案，读取保存响应中的版本号，再将它作为下一次保存的期望版本。

## 错误信息

```text
first response version=7, want committed version 8
```

