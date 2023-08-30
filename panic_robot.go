package monitor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

type (
	PanicRobotOption func(*panicRobotOption)

	panicRobotOption struct {
		Enable        bool
		env           string
		wechatEnable  bool
		wechatPushUrl string
		feishuEnable  bool
		feishuPushUrl string
	}

	PanicRobot struct {
		Env      string
		HostName string
		Wechat   robotConfig
		Feishu   robotConfig
	}

	robotConfig struct {
		Enable  bool
		PushUrl string
	}
)

func PanicRobotEnable(enable bool) PanicRobotOption {
	return func(o *panicRobotOption) {
		o.Enable = enable
	}
}

func PanicRobotEnv(env string) PanicRobotOption {
	return func(o *panicRobotOption) {
		o.env = env
	}
}

func PanicRobotWechatEnable(enable bool) PanicRobotOption {
	return func(o *panicRobotOption) {
		o.wechatEnable = enable
	}
}

func PanicRobotWechatPushUrl(url string) PanicRobotOption {
	return func(o *panicRobotOption) {
		o.wechatPushUrl = url
	}
}

func PanicRobotFeishuEnable(enable bool) PanicRobotOption {
	return func(o *panicRobotOption) {
		o.feishuEnable = enable
	}
}

func PanicRobotFeishuPushUrl(url string) PanicRobotOption {
	return func(o *panicRobotOption) {
		o.feishuPushUrl = url
	}
}

func NewPanicRobot(opts ...PanicRobotOption) (*PanicRobot, error) {
	opt := &panicRobotOption{}

	for _, f := range opts {
		f(opt)
	}

	if !opt.Enable {
		return nil, errors.New("PanicRobot Disable")
	}

	if opt.wechatEnable && opt.wechatPushUrl == "" {
		return nil, errors.New("Wechat push url Can not be Null ")
	}

	if opt.feishuEnable && opt.feishuPushUrl == "" {
		return nil, errors.New("Feishu push url Can not be Null ")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &PanicRobot{
		Env:      opt.env,
		HostName: hostname,
		Wechat: robotConfig{
			Enable:  opt.wechatEnable,
			PushUrl: opt.wechatPushUrl,
		},
		Feishu: robotConfig{
			Enable:  opt.feishuEnable,
			PushUrl: opt.feishuPushUrl,
		},
	}, nil
}

func (pr *PanicRobot) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				env := spliceStr("Env: ", pr.Env, "\n")
				host := spliceStr("Host: ", pr.HostName, "\n")
				datetime := spliceStr("Time: ", time.Now().Format("2006-01-02 15:04:05"), "\n")
				request := spliceStr("Request: ", c.Request.Method, " ", c.Request.URL.Path, "\n")
				params := spliceStr("Params: ", c.Request.URL.RawQuery, "\n")
				msg := spliceStr("Panic: ", fmt.Sprintf("%v", err), "\n")

				bodyRaw, _ := io.ReadAll(c.Request.Body)
				body := spliceStr("Body: ", string(bodyRaw), "\n")

				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				stack := string(buf[:n])

				content := spliceStr(env, host, datetime, request, params, body, msg, stack)

				go pr.wechatPush(content)
				go pr.feishuPush(content)

				c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"code":  500,
					"msg":   "Server Error",
					"data":  nil,
					"trace": err,
				})
			}
		}()

		c.Next()
	}
}

func (pr *PanicRobot) wechatPush(content string) {
	if pr.Wechat.Enable {
		body := map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": content,
			},
		}

		push(body, pr.Wechat.PushUrl)
	}
}

func (pr *PanicRobot) feishuPush(content string) {
	if pr.Feishu.Enable {
		body := map[string]interface{}{
			"msg_type": "text",
			"content": map[string]string{
				"text": content,
			},
		}

		push(body, pr.Feishu.PushUrl)
	}
}

func push(content map[string]interface{}, pushUrl string) {
	contentBytes, err := json.Marshal(content)
	if err != nil {
		log.Println("Marshal push content failed:", err.Error())
	}

	_, err = http.Post(pushUrl, "application/json; charset=utf-8", bytes.NewBuffer(contentBytes))
	if err != nil {
		log.Println("Post push content failed:", err.Error(), " pushUrl:", pushUrl)
	}
}
