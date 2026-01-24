# NapCat 配置指南

NapCat 是基于 QQNT 协议的 QQ Bot 框架，无需签名服务器，兼容 go-cqhttp API。

---

## 1. 前置条件

安装最新版 **QQNT 客户端**（官方 QQ）：
- https://im.qq.com/pcqq

---

## 2. 下载 NapCat

从 GitHub 下载：
- https://github.com/NapNeko/NapCatQQ/releases

下载 `NapCat.Shell.zip` 并解压到任意目录（如 `D:\NapCat`）

---

## 3. 启动 NapCat

```powershell
cd D:\NapCat
.\napcat-utf8.bat
```

首次运行会自动打开 QQ 登录界面，扫码登录。

---

## 4. 配置 WebSocket

1. 打开 NapCat WebUI：http://127.0.0.1:6099

2. 进入「网络配置」

3. 添加「正向 WebSocket 服务」：
   - 名称：`ws-server`
   - 监听地址：`127.0.0.1`
   - 端口：`6700`
   - 启用：✅

4. 保存并重启 NapCat

---

## 5. 运行 QQ Bot

确保 NapCat 正在运行，然后：

```powershell
cd d:\file\MemoryOs
go run examples/qqbot/main_real.go
```

或者使用编译好的版本：
```powershell
.\examples\qqbot\qqbot_real.exe
```

---

## 6. 测试

用另一个 QQ 号私聊你的 Bot QQ，应该能收到回复！

---

## 常见问题

### Q: 连接失败 "dial tcp 127.0.0.1:6700"
- 确保 NapCat 正在运行
- 确保已在 WebUI 配置正向 WebSocket

### Q: NapCat 启动报错
- 确保已安装最新版 QQNT
- 使用 `napcat-utf8.bat` 启动

### Q: 消息无回复
- 检查 MemoryOS 配置（config/config.yaml）的 API Key
- 查看终端日志是否有错误

### Q: QQ 被风控
- 新号需要养几天
- 回复频率不要太快
- 建议用小号测试
