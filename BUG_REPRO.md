# Bug Reproduction

## Bug 是什么

草稿从模板 tv1 切换到 tv2 后，回滚检查点记录成了切换后的模板，无法恢复切换前的来源模板。

## 如何触发

创建使用 tv1 的草稿，将它切换到 tv2，然后读取本次切换产生的回滚检查点。

## 错误信息

```text
checkpoint captured template "tv2", want source tv1
```

