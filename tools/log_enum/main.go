package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ikuaisdk "github.com/zy84338719/ikuai-api"
)

func main() {
	ctx := context.Background()

	client, err := ikuaisdk.NewClientWithLogin("10.10.30.254", "zhangyi", "zx19950124", ikuaisdk.WithTimeout(30*time.Second))
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Printf("登录成功，版本: %s\n\n", client.GetVersion())

	logFuncs := []string{
		"syslog",
		"log",
		"warning_msg_log",
		"warning_msg",
		"warning_log",
		"alert_log",
		"alert_msg",
		"syslog_warning",
		"syslog_alert",
		"log_warning",
		"log_alert",
		"notice",
		"syslog_notice",
		"syslog-notice",
		"syslog_warning",
		"syslog_alert",
		"syslog-wanpppoe",
		"syslog-dhcpd",
		"syslog-arp",
		"syslog-ddns",
		"syslog-webadmin",
		"syslog-sysevent",
		"system_log",
		"security_log",
	}

	for _, fn := range logFuncs {
		fmt.Printf("=== 尝试 %s ===\n", fn)

		var resp map[string]interface{}
		err := client.Call(ctx, fn, "show", nil, &resp)
		if err != nil {
			fmt.Printf("错误: %v\n\n", err)
			continue
		}

		if code, ok := resp["code"].(float64); ok && code == 0 {
			data, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Printf("✅ 成功!\n%s\n\n", string(data))
		} else {
			msg, _ := resp["message"].(string)
			fmt.Printf("失败: %s\n\n", msg)
		}
	}
}
