# Bug Reproduction

## Bug 是什么

向访客分享一个附件后，同一文档下没有分享的其他附件也能被该访客打开。

## 如何触发

在同一文档中创建两个附件，只向访客分享其中一个，然后让该访客打开另一个附件。

## 错误信息

```text
opening unshared sibling returned <nil>, want forbidden
```

