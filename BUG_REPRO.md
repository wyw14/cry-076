# Bug Reproduction

## Bug 是什么

导出文件已经写入存储后，如果导出结果登记失败，清理流程没有删除实际存储的文件，遗留了孤立产物。

## 如何触发

让文件存储成功返回 `exports/export-test-b`，再让结果登记返回错误，并检查清理时请求删除的路径。

## 错误信息

```text
deleted path="export-test-b", want stored path "exports/export-test-b"
```
