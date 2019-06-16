# hustoj 压测小工具

需要开启 web 的 `benchmark mode`。

## 0.为啥输出都是英文的

这样会显得工具非常专业。

## 1.如何使用

```shell
./timetest -h
Usage of ./timetest:
  -c int
        concurrency (default 10)
  -code string
        submit code file path (default "./1000.c")
  -config string
        config file path (default "./config.json")
  -n int
        total (default 100)
```

## 2.输出都有啥

```text
./timetest -c 2 -code ./example/1000.c -config ./example/config.json -n 10                                                                                                   2 ↵
stage - submit
-----
average: 245.34873ms // 你懂得8
pct90: 305.141436ms // 耗时从少到多排序，处在 90% 这个位置的提交耗时
pct95: 305.141436ms // 同理易得
pct99: 305.141436ms

10 valid result
total cost -> 1227ms // 总计的提交
ok 10 // 跑分模式下返回了submit id的提交数量
50x 0  // web顶不住了返回了50x的数量
unknown 0 // 其他情况，比如本应 keep alive 的链接没有了
qps 8.147686 (valid value only) // qps 仅计算返回了submit id的情况 实际会比这个数字高一点
```

## 3.附言

`url` 结尾有个斜杠，是必须的。

我的小服务器有点弱，测试请轻点。
