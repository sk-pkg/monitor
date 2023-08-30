# PanicRobot 

PanicRobot是适用于gin框架的一个`panic`监控推送处理包，支持企业微信和飞书群机器人推送。

## 安装

```shell
go get -u "github.com/sk-pkg/monitor"
```

## 快速开始

### 配置文件
/bin/configs/{env}.json
```json
 "monitor": {
    "panic_robot": {
      "enable": true,
      "wechat": {
        "enable": true,
        "push_url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxxx"
      },
      "feishu": {
        "enable": true,
        "push_url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxxx"
      }
    }
  },
```
/app/config.go
```go
Monitor struct {
	PanicRobot PanicRobot `json:"panic_robot"`
}

PanicRobot struct {
    Enable bool        `json:"enable"`
	Wechat robotConfig `json:"wechat"`
	Feishu robotConfig `json:"feishu"`
}

robotConfig struct {
	Enable  bool   `json:"enable"`
	PushUrl string `json:"push_url"`
}
```

### 开始使用
/bootstrap/http.go
```go
package bootstrap

func (a *App) loadMux() {
    mux.Use(gin.Recovery())
    a.loadPanicRobot(mux)	// panic监控
    a.Mux = mux
    a.Logger.Info("Mux loaded successfully")
}

// loadPanicRobot 加载panic监控机器人
func (a *App) loadPanicRobot(mux *gin.Engine) {
    panicRobot, err := monitor.NewPanicRobot(
		monitor.PanicRobotEnable(os.Getenv(a.Config.System.EnvKey)),
        monitor.PanicRobotEnv(a.Config.Monitor.PanicRobot.Enable),
        monitor.PanicRobotWechatEnable(a.Config.Monitor.PanicRobot.Wechat.Enable),
        monitor.PanicRobotWechatPushUrl(a.Config.Monitor.PanicRobot.Wechat.PushUrl),
        monitor.PanicRobotFeishuEnable(a.Config.Monitor.PanicRobot.Feishu.Enable),
        monitor.PanicRobotFeishuPushUrl(a.Config.Monitor.PanicRobot.Feishu.PushUrl),
    )

    if err == nil {
        mux.Use(panicRobot.Middleware())
    }
}
```
