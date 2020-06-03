# 小爱拦截器
### 编译
    执行:
    env GOOS=linux GOARCH=arm GOARM=5 go build

### 说明
+ 使用fsnotify监控变动实时触发
+ 内置简单jsonrpc服务器，可以控制播放,获取状态
+ 采用异步模式方便编写hass自动化

### 使用
+ /root/xiaoai_hook -url 服务器地址 -token rpc令牌
+ token为可选参数，使用时调用jsonrpc加入相同的token 
+ /root/xiaoai_hook -s stop 结束进程
+ node-red-demo.json 为node-red调用样例

