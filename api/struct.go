package api

import "fmt"

// IKuaiVersion represents detected iKuai OS version
type IKuaiVersion int

const (
	VersionUnknown IKuaiVersion = iota
	VersionV3
	VersionV4
)

// String returns string representation of version
func (v IKuaiVersion) String() string {
	switch v {
	case VersionV3:
		return "v3"
	case VersionV4:
		return "v4"
	default:
		return "unknown"
	}
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"passwd"`
	Pass     string `json:"pass"`
}

type LoginResp struct {
	CallResp
}

// IsV4 detects if response is from v4 API
func (r *LoginResp) IsV4() bool {
	return r.Message != ""
}

type CallReq struct {
	FuncName string      `json:"func_name"`
	Action   string      `json:"action"`
	Param    interface{} `json:"param,omitempty"`
}

type CallResp struct {
	ErrMsg  string `json:"ErrMsg"`
	Message string `json:"message"` // v4 format
	Result  int    `json:"Result"`
	Code    int    `json:"code"` // v4 format
}

// IsSuccess returns true if API call was successful (v3 or v4)
// v3: Result=10000 (login) or Result=30000 (call)
// v4: Code=0
func (c *CallResp) IsSuccess() bool {
	// Check for v4 format first (code field present and non-zero, or code=0 with message)
	if c.Code != 0 {
		return c.Code == 0
	}
	// If code is present but zero, check for v4 indicators
	if c.Message != "" {
		return c.Code == 0
	}
	// v3 format
	return c.Result == 10000 || c.Result == 30000
}

// IsLoginFailed returns true if login failed
// v3: Result=10014
// v4: Code != 0
func (c *CallResp) IsLoginFailed() bool {
	if c.Code != 0 {
		return c.Code != 0
	}
	if c.Message != "" {
		return c.Code != 0
	}
	return c.Result == 10014
}

// GetErrMsg returns error message, supporting both v3 and v4 formats
func (c *CallResp) GetErrMsg() string {
	if c.ErrMsg != "" {
		return c.ErrMsg
	}
	if c.Message != "" {
		return c.Message
	}
	if c.Code != 0 {
		return fmt.Sprintf("Error code: %d", c.Code)
	}
	return ""
}

// GetResult returns result code for compatibility with existing code
// This is deprecated in favor of IsSuccess() but kept for backward compatibility
func (c *CallResp) GetResult() int {
	if c.Code != 0 {
		if c.Code == 0 {
			return 30000 // v3 equivalent for v4 success
		}
		return c.Code
	}
	return c.Result
}

type WebUserShowResp struct {
	CallResp
}

type CustomISPShowResp struct {
	CallResp
	Data struct {
		Total int `json:"total"`
		Data  []struct {
			ID      int    `json:"id"`
			Name    string `json:"name"`
			IPGroup string `json:"ipgroup"`
			Comment string `json:"comment"`
			Time    string `json:"time"`
		} `json:"data"`
	} `json:"Data"`
	// v4 format uses "results" instead of "Data"
	Results *struct {
		Data []struct {
			ID      int    `json:"id"`
			Name    string `json:"name"`
			IPGroup string `json:"ipgroup"`
			Comment string `json:"comment"`
			Time    string `json:"time"`
		} `json:"data"`
	} `json:"results,omitempty"`
}

// GetData returns data from either Data (v3) or Results (v4)
func (c *CustomISPShowResp) GetData() []interface{} {
	if c.Results != nil && c.Results.Data != nil {
		var result []interface{}
		for _, v := range c.Results.Data {
			result = append(result, v)
		}
		return result
	}
	var result []interface{}
	for _, v := range c.Data.Data {
		result = append(result, v)
	}
	return result
}

type CustomISPDelResp struct {
	CallResp
}

type CustomISPAddResp struct {
	CallResp
}

type StreamDomainShowResp struct {
	CallResp
	Data struct {
		Total int `json:"total"`
		Data  []struct {
			ID        int    `json:"id"`
			Interface string `json:"interface"`
			SrcAddr   string `json:"src_addr"`
			Enabled   string `json:"enabled"`
			Week      string `json:"week"`
			Comment   string `json:"comment"`
			Domain    string `json:"domain"`
			Time      string `json:"time"`
		} `json:"data"`
	} `json:"Data"`
	// v4 format uses "results" instead of "Data"
	Results *struct {
		Data []struct {
			ID        int    `json:"id"`
			Interface string `json:"interface"`
			SrcAddr   string `json:"src_addr"`
			Enabled   string `json:"enabled"`
			Week      string `json:"week"`
			Comment   string `json:"comment"`
			Domain    string `json:"domain"`
			Time      string `json:"time"`
		} `json:"data"`
	} `json:"results,omitempty"`
}

// GetData returns data from either Data (v3) or Results (v4)
func (c *StreamDomainShowResp) GetData() []interface{} {
	if c.Results != nil && c.Results.Data != nil {
		var result []interface{}
		for _, v := range c.Results.Data {
			result = append(result, v)
		}
		return result
	}
	var result []interface{}
	for _, v := range c.Data.Data {
		result = append(result, v)
	}
	return result
}

type StreamDomainDelResp struct {
	CallResp
}

type StreamDomainAddResp struct {
	CallResp
}

type HomepageShowSysStatResp struct {
	CallResp
	Data struct {
		SysStat struct {
			Cpu        []string `json:"cpu"`
			CpuTemp    []int    `json:"cputemp"`
			Freq       []string `json:"freq"`
			GWid       string   `json:"gwid"`
			Hostname   string   `json:"hostname"`
			LinkStatus int      `json:"link_status"`
			Memory     struct {
				Total     int64  `json:"total"`
				Available int64  `json:"available"`
				Free      int64  `json:"free"`
				Cached    int64  `json:"cached"`
				Buffers   int64  `json:"buffers"`
				Used      string `json:"used"`
			} `json:"memory"`
			OnlineUser struct {
				Count         int `json:"count"`
				Count2G       int `json:"count_2g"`
				Count5G       int `json:"count_5g"`
				CountWired    int `json:"count_wired"`
				CountWireless int `json:"count_wireless"`
			} `json:"online_user"`
			Stream struct {
				ConnectNum int   `json:"connect_num"`
				Upload     int   `json:"upload"`
				Download   int   `json:"download"`
				TotalUp    int64 `json:"total_up"`
				TotalDown  int64 `json:"total_down"`
			} `json:"stream"`
			Uptime  int `json:"uptime"`
			VerInfo struct {
				ModelName    string `json:"modelname"`
				VerString    string `json:"verstring"`
				Version      string `json:"version"`
				BuildDate    int64 `json:"build_date"`
				Arch         string `json:"arch"`
				SysBit       string `json:"sysbit"`
				VerFlags     string `json:"verflags"`
				IsEnterprise int    `json:"is_enterprise"`
				SupportI18N  int    `json:"support_i18n"`
				SupportLcd   int    `json:"support_lcd"`
			} `json:"verinfo"`
		} `json:"sysstat"`
		AcStatus struct {
			ApCount  int `json:"ap_count"`
			ApOnline int `json:"ap_online"`
		} `json:"ac_status"`
	} `json:"Data"`
	// v4 format uses "results" instead of "Data"
	Results *struct {
		SysStat struct {
			Cpu        []string `json:"cpu"`
			CpuTemp    []int    `json:"cputemp"`
			Freq       []string `json:"freq"`
			GWid       string   `json:"gwid"`
			Hostname   string   `json:"hostname"`
			LinkStatus int      `json:"link_status"`
			Memory     struct {
				Total     int64  `json:"total"`
				Available int64  `json:"available"`
				Free      int64  `json:"free"`
				Cached    int64  `json:"cached"`
				Buffers   int64  `json:"buffers"`
				Used      string `json:"used"`
			} `json:"memory"`
			OnlineUser struct {
				Count         int `json:"count"`
				Count2G       int `json:"count_2g"`
				Count5G       int `json:"count_5g"`
				CountWired    int `json:"count_wired"`
				CountWireless int `json:"count_wireless"`
			} `json:"online_user"`
			Stream struct {
				ConnectNum int   `json:"connect_num"`
				Upload     int   `json:"upload"`
				Download   int   `json:"download"`
				TotalUp    int64 `json:"total_up"`
				TotalDown  int64 `json:"total_down"`
			} `json:"stream"`
			Uptime  int `json:"uptime"`
			VerInfo struct {
				ModelName    string `json:"modelname"`
				VerString    string `json:"verstring"`
				Version      string `json:"version"`
				BuildDate    int64 `json:"build_date"`
				Arch         string `json:"arch"`
				SysBit       string `json:"sysbit"`
				VerFlags     string `json:"verflags"`
				IsEnterprise int    `json:"is_enterprise"`
				SupportI18N  int    `json:"support_i18n"`
				SupportLcd   int    `json:"support_lcd"`
			} `json:"verinfo"`
		} `json:"sysstat"`
		AcStatus struct {
			ApCount  int `json:"ap_count"`
			ApOnline int `json:"ap_online"`
		} `json:"ac_status"`
	} `json:"results,omitempty"`
}

// GetData returns data from either Data (v3) or Results (v4)
func (c *HomepageShowSysStatResp) GetData() interface{} {
	if c.Results != nil {
		return *c.Results
	}
	return c.Data
}

type MonitorLanIPShowResp struct {
	CallResp
	Data struct {
		Data []struct {
			ApName       string `json:"apname"`
			AcGid        int    `json:"ac_gid"`
			Mac          string `json:"mac"`
			LinkAddr     string `json:"link_addr"`
			Hostname     string `json:"hostname"`
			DTalkName    string `json:"dtalk_name"`
			DownRate     string `json:"downrate"`
			Reject       int    `json:"reject"`
			Uprate       string `json:"uprate"`
			Signal       interface{} `json:"signal"`
			ClientType   string `json:"client_type"`
			Bssid        string `json:"bssid"`
			AuthType     int    `json:"auth_type"`
			WebID        int    `json:"webid"`
			Comment      string `json:"comment"`
			Username     string `json:"username"`
			PPPType      string `json:"ppptype"`
			ApMac        string `json:"apmac"`
			Upload       int    `json:"upload"`
			Ssid         string `json:"ssid"`
			Frequencies  string `json:"frequencies"`
			Uptime       string `json:"uptime"`
			Id           int    `json:"id"`
			IpAddrInt    int64  `json:"ip_addr_int"`
			ConnectNum   int    `json:"connect_num"`
			IpAddr       string `json:"ip_addr"`
			Download     int    `json:"download"`
			TotalUp      int64  `json:"total_up"`
			TotalDown    int64  `json:"total_down"`
			ClientDevice string `json:"client_device"`
			Timestamp    int    `json:"timestamp"`
		} `json:"data"`
	} `json:"Data"`
	// v4 format uses "results" instead of "Data"
	Results *struct {
		Data []struct {
			ApName       string `json:"apname"`
			AcGid        int    `json:"ac_gid"`
			Mac          string `json:"mac"`
			LinkAddr     string `json:"link_addr"`
			Hostname     string `json:"hostname"`
			DTalkName    string `json:"dtalk_name"`
			DownRate     string `json:"downrate"`
			Reject       int    `json:"reject"`
			Uprate       string `json:"uprate"`
			Signal       interface{} `json:"signal"`
			ClientType   string `json:"client_type"`
			Bssid        string `json:"bssid"`
			AuthType     int    `json:"auth_type"`
			WebID        int    `json:"webid"`
			Comment      string `json:"comment"`
			Username     string `json:"username"`
			PPPType      string `json:"ppptype"`
			ApMac        string `json:"apmac"`
			Upload       int    `json:"upload"`
			Ssid         string `json:"ssid"`
			Frequencies  string `json:"frequencies"`
			Uptime       string `json:"uptime"`
			Id           int    `json:"id"`
			IpAddrInt    int64  `json:"ip_addr_int"`
			ConnectNum   int    `json:"connect_num"`
			IpAddr       string `json:"ip_addr"`
			Download     int    `json:"download"`
			TotalUp      int64  `json:"total_up"`
			TotalDown    int64  `json:"total_down"`
			ClientDevice string `json:"client_device"`
			Timestamp    int    `json:"timestamp"`
		} `json:"data"`
	} `json:"results,omitempty"`
}

// GetData returns data from either Data (v3) or Results (v4)
func (c *MonitorLanIPShowResp) GetData() []interface{} {
	if c.Results != nil && c.Results.Data != nil {
		var result []interface{}
		for _, v := range c.Results.Data {
			result = append(result, v)
		}
		return result
	}
	var result []interface{}
	for _, v := range c.Data.Data {
		result = append(result, v)
	}
	return result
}

type MonitorIFaceShowResp struct {
	CallResp
	Data struct {
		IFaceCheck []struct {
			Id              int    `json:"id"`
			Interface       string `json:"interface"`
			ParentInterface string `json:"parent_interface"`
			IpAddr          string `json:"ip_addr"`
			Gateway         string `json:"gateway"`
			Internet        string `json:"internet"`
			UpdateTime      string `json:"updatetime"`
			AutoSwitch      string `json:"auto_switch"`
			Result          string `json:"result"`
			ErrMsg          string `json:"errmsg"`
			Comment         string `json:"comment"`
		} `json:"iface_check"`
		IFaceStream []struct {
			Interface   string `json:"interface"`
			Comment     string `json:"comment"`
			IpAddr      string `json:"ip_addr"`
			ConnectNum  string `json:"connect_num"`
			Upload      int    `json:"upload"`
			Download    int    `json:"download"`
			TotalUp     int64 `json:"total_up"`
			TotalDown   int64 `json:"total_down"`
			UpDropped   int    `json:"updropped"`
			DownDropped int    `json:"downdropped"`
			UpPacked    int    `json:"uppacked"`
			DownPacked  int    `json:"downpacked"`
		} `json:"iface_stream"`
	} `json:"Data"`
	// v4 format uses "results" instead of "Data"
	Results *struct {
		IFaceCheck []struct {
			Id              int    `json:"id"`
			Interface       string `json:"interface"`
			ParentInterface string `json:"parent_interface"`
			IpAddr          string `json:"ip_addr"`
			Gateway         string `json:"gateway"`
			Internet        string `json:"internet"`
			UpdateTime      string `json:"updatetime"`
			AutoSwitch      string `json:"auto_switch"`
			Result          string `json:"result"`
			ErrMsg          string `json:"errmsg"`
			Comment         string `json:"comment"`
			Signal         interface{} `json:"signal"`
		} `json:"iface_check"`
		IFaceStream []struct {
			Interface   string `json:"interface"`
			Comment     string `json:"comment"`
			IpAddr      string `json:"ip_addr"`
			ConnectNum  string `json:"connect_num"`
			Upload      int    `json:"upload"`
			Download    int    `json:"download"`
			TotalUp     int64 `json:"total_up"`
			TotalDown   int64 `json:"total_down"`
			UpDropped   int    `json:"updropped"`
			DownDropped int    `json:"downdropped"`
			UpPacked    int    `json:"uppacked"`
			DownPacked  int    `json:"downpacked"`
		} `json:"iface_stream"`
	} `json:"results,omitempty"`
}

// GetData returns data from either Data (v3) or Results (v4)
func (c *MonitorIFaceShowResp) GetData() interface{} {
	if c.Results != nil {
		return *c.Results
	}
	return c.Data
}
