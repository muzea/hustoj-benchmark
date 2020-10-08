## 食用方式

从 `releases` 中下载所需的版本，运行可执行文件，浏览器访问 `http://localhost:8000/`,修改服务器配置为自己实际的配置，并在 `hustoj` 中开启跑分模式，即可开始测试。

## 开启跑分模式

将 [这个变量](https://github.com/zhblue/hustoj/blob/6bd0c33455b82ffa1e8856f4f1f48501cf2828da/trunk/web/include/db_info.inc.php#L67) 改为 `true`

## 目前存在的问题

- `websocket` 没有处理异常，会经常导致崩溃
- 没有对异常输入做校验，需要自己注意
- 计数器加减不是原子的，偶尔会出现计数丢失几个
