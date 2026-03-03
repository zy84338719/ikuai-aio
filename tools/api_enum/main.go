package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	username   string
	password   string
}

func NewClient(baseURL, username, password string) *Client {
	baseURL = strings.TrimSpace(baseURL)
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Login() error {
	salt := "salt_11"
	b64Pass := base64Encode(salt + c.password)

	req := map[string]interface{}{
		"username": c.username,
		"passwd":   md5Hash(c.password),
		"pass":     b64Pass,
	}

	resp, err := c.post("/Action/login", req)
	if err != nil {
		return err
	}

	if resp["code"] != nil && resp["code"].(float64) != 0 {
		return fmt.Errorf("login failed: %v", resp["message"])
	}
	if resp["Result"] != nil && resp["Result"].(float64) != 10000 {
		return fmt.Errorf("login failed: %v", resp["ErrMsg"])
	}

	return nil
}

func (c *Client) Call(funcName, action string, param map[string]interface{}) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"func_name": funcName,
		"action":    action,
	}
	if param != nil {
		req["param"] = param
	}
	return c.post("/Action/call", req)
}

func (c *Client) post(path string, body interface{}) (map[string]interface{}, error) {
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", c.baseURL+path, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func md5Hash(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func base64Encode(s string) string {
	return strings.TrimRight(
		strings.ReplaceAll(
			strings.ReplaceAll(
				url.QueryEscape(s),
				"%", "",
			),
			"=", "",
		),
		"+",
	)
}

func isSuccess(resp map[string]interface{}) bool {
	if resp["code"] != nil {
		return resp["code"].(float64) == 0
	}
	if resp["Result"] != nil {
		r := resp["Result"].(float64)
		return r == 10000 || r == 30000
	}
	return false
}

func getErrorMsg(resp map[string]interface{}) string {
	if resp["message"] != nil {
		return resp["message"].(string)
	}
	if resp["ErrMsg"] != nil {
		return resp["ErrMsg"].(string)
	}
	return ""
}

func prettyJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

var funcNames = []string{
	"homepage",
	"monitor_lanip",
	"monitor_lanipv6",
	"monitor_iface",
	"monitor_iface_check",
	"monitor_stream",
	"monitor_conn",
	"monitor_sysstat",
	"monitor_disk",
	"monitor_arp",
	"system",
	"sysuser",
	"webuser",
	"network",
	"interface",
	"wan",
	"lan",
	"dhcp",
	"dns",
	"static_route",
	"policy_route",
	"nat",
	"firewall",
	"acl",
	"ipgroup",
	"macgroup",
	"timegroup",
	"domain_group",
	"custom_isp",
	"stream_domain",
	"dns_redirect",
	"upnp",
	"ddns",
	"vpn_l2tp",
	"vpn_pptp",
	"vpn_ipsec",
	"vpn_openvpn",
	"vpn_wireguard",
	"qos",
	"flow_distribute",
	"protocol_group",
	"app_filter",
	"url_filter",
	"web_filter",
	"time_limit",
	"offline_limit",
	"port_map",
	"dmz",
	"alias_ip",
	"vlan",
	"bonding",
	"bridge",
	"wireless",
	"ap",
	"ac",
	"remote_manage",
	"backup",
	"upgrade",
	"reboot",
	"reset",
	"log",
	"system_log",
	"security_log",
	"flow_log",
	"online_log",
	"dhcp_log",
	"vpn_log",
	"radius",
	"portal",
	"bind_mac",
	"bind_ip",
	"arp_bind",
	"behavior",
	"app_behavior",
	"web_behavior",
	"speed_limit",
	"conn_limit",
	"bandwidth",
	"multi_wan",
	"load_balance",
	"failover",
	"iptv",
	"voip",
	"ipv6",
	"radvd",
	"dns64",
	"nat64",
	"snat",
	"dnat",
	"masq",
	"proxy",
	"http_proxy",
	"socks_proxy",
	"cache",
	"ad_filter",
	"adblock",
	"dns_filter",
	"domain_filter",
	"keyword_filter",
	"user_group",
	"auth",
	"pppoe",
	"pptp_client",
	"l2tp_client",
	"openvpn_client",
	"wireguard_client",
	"certificate",
	"ssl",
	"ssh",
	"telnet",
	"snmp",
	"syslog",
	"netflow",
	"sflow",
	"prometheus",
	"grafana",
	"cloud",
	"cloud_login",
	"quick_setup",
	"wizard",
	"language",
	"theme",
	"help",
	"about",
	"license",
	"plugin",
	"module",
	"service",
	"crontab",
	"schedule",
	"task",
	"notification",
	"alert",
	"email",
	"webhook",
	"telegram",
	"wechat",
	"dingtalk",
}

func main() {
	client := NewClient("10.10.30.254", "zhangyi", "zx19950124")

	fmt.Println("=== 连接 iKuai 路由器 ===")
	fmt.Println("地址: 10.10.30.254")
	fmt.Println("用户: zhangyi")
	fmt.Println()

	if err := client.Login(); err != nil {
		fmt.Printf("登录失败: %v\n", err)
		return
	}
	fmt.Println("登录成功!")
	fmt.Println()

	successAPIs := []string{}
	failedAPIs := []string{}
	apiDetails := make(map[string]map[string]interface{})

	actions := []string{"show", "list", "get"}

	fmt.Println("=== 开始枚举 API 接口 ===")
	fmt.Println()

	for _, funcName := range funcNames {
		found := false
		for _, action := range actions {
			resp, err := client.Call(funcName, action, nil)
			if err != nil {
				continue
			}

			if isSuccess(resp) {
				found = true
				successAPIs = append(successAPIs, funcName)
				apiDetails[funcName] = resp
				fmt.Printf("[OK] %s (%s)\n", funcName, action)
				break
			}
		}

		if !found {
			failedAPIs = append(failedAPIs, funcName)
		}
	}

	fmt.Println()
	fmt.Println("=== 可用的 API 接口详情 ===")
	fmt.Println()

	for _, api := range successAPIs {
		fmt.Printf("--- %s ---\n", api)
		fmt.Println(prettyJSON(apiDetails[api]))
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("=== 统计 ===")
	fmt.Printf("成功: %d 个\n", len(successAPIs))
	fmt.Printf("失败: %d 个\n", len(failedAPIs))
	fmt.Println()

	fmt.Println("=== 可用接口列表 ===")
	for _, api := range successAPIs {
		fmt.Printf("  - %s\n", api)
	}
}
