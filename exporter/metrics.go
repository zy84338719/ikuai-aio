package exporter

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/NERVEbing/ikuai-aio/api"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	client *api.Client

	version *prometheus.Desc
	up      *prometheus.Desc
	uptime  *prometheus.Desc

	cpuUsageRatio  *prometheus.Desc
	cpuTemperature *prometheus.Desc

	memorySizeKiloBytes    *prometheus.Desc
	memoryUsageKiloBytes   *prometheus.Desc
	memoryCachedKiloBytes  *prometheus.Desc
	memoryBuffersKiloBytes *prometheus.Desc

	interfaceInfo *prometheus.Desc

	deviceCount *prometheus.Desc
	deviceInfo  *prometheus.Desc

	networkUploadTotalBytes   *prometheus.Desc
	networkDownloadTotalBytes *prometheus.Desc
	networkUploadSpeedBytes   *prometheus.Desc
	networkDownloadSpeedBytes *prometheus.Desc
	networkConnectCount       *prometheus.Desc
}

func NewMetrics(namespace string) *Metrics {
	client := api.NewClient()
	if err := client.Login(); err != nil {
		log.Fatalln(err)
	}

	return &Metrics{
		client:                    client,
		version:                   newDesc(namespace, "version", "", []string{"version", "arch", "ver_string"}),
		up:                        newDesc(namespace, "up", "", []string{"id"}),
		uptime:                    newDesc(namespace, "uptime", "", []string{"id"}),
		cpuUsageRatio:             newDesc(namespace, "cpu_usage_ratio", "", []string{"id"}),
		cpuTemperature:            newDesc(namespace, "cpu_temperature", "", nil),
		memorySizeKiloBytes:       newDesc(namespace, "memory_size_kilo_bytes", "", nil),
		memoryUsageKiloBytes:      newDesc(namespace, "memory_usage_kilo_bytes", "", nil),
		memoryCachedKiloBytes:     newDesc(namespace, "memory_cached_kilo_bytes", "", nil),
		memoryBuffersKiloBytes:    newDesc(namespace, "memory_buffers_kilo_bytes", "", nil),
		interfaceInfo:             newDesc(namespace, "interface_info", "", []string{"id", "interface", "comment", "internet", "parent_interface", "ip_addr", "display"}),
		deviceCount:               newDesc(namespace, "device_count", "", nil),
		deviceInfo:                newDesc(namespace, "device_info", "", []string{"id", "mac", "hostname", "ip_addr", "comment", "display"}),
		networkUploadTotalBytes:   newDesc(namespace, "network_upload_total_bytes", "", []string{"id", "display", "ip_addr"}),
		networkDownloadTotalBytes: newDesc(namespace, "network_download_total_bytes", "", []string{"id", "display", "ip_addr"}),
		networkUploadSpeedBytes:   newDesc(namespace, "network_upload_speed_bytes", "", []string{"id", "display", "ip_addr"}),
		networkDownloadSpeedBytes: newDesc(namespace, "network_download_speed_bytes", "", []string{"id", "display", "ip_addr"}),
		networkConnectCount:       newDesc(namespace, "network_connect_count", "", []string{"id", "display", "ip_addr"}),
	}
}

func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.version
	ch <- m.up
	ch <- m.uptime
	ch <- m.cpuUsageRatio
	ch <- m.cpuTemperature
	ch <- m.memorySizeKiloBytes
	ch <- m.memoryUsageKiloBytes
	ch <- m.memoryCachedKiloBytes
	ch <- m.memoryBuffersKiloBytes
	ch <- m.interfaceInfo
	ch <- m.deviceCount
	ch <- m.deviceInfo
	ch <- m.networkUploadTotalBytes
	ch <- m.networkDownloadTotalBytes
	ch <- m.networkUploadSpeedBytes
	ch <- m.networkDownloadSpeedBytes
	ch <- m.networkConnectCount
}

func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	defer func() {
		if err := recover(); err != nil {
			logger("recover", "error: %s", err)
			ch <- prometheus.MustNewConstMetric(
				m.up,
				prometheus.GaugeValue,
				0,
				"host",
			)
		}
	}()

	if !m.client.IsLogin() {
		logger("IsLogin", "cookie has expired, try logged in again")
		if err := m.client.Login(); err != nil {
			logger("Login", "error: %s", err)
			return
		}
		logger("Login", "success")
	}

	homepageShowSysStatResp, err := m.client.HomepageShowSysStat()
	if err != nil {
		logger("HomepageShowSysStat", "error: %s", err)
		return
	}
	monitorLanIPShowResp, err := m.client.MonitorLanIPShow()
	if err != nil {
		logger("MonitorLanIPShow", "error: %s", err)
		return
	}
	monitorIFaceShowResp, err := m.client.MonitorIFaceShow()
	if err != nil {
		logger("MonitorIFaceShow", "error: %s", err)
		return
	}

	// Handle both v3 and v4 formats
	var sysStat struct {
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
	}

	// Get sysstat data - handle both v3 and v4 formats
	if homepageShowSysStatResp.Results != nil {
		sysStat = homepageShowSysStatResp.Results.SysStat
	} else {
		sysStat = homepageShowSysStatResp.Data.SysStat
	}

	// Get iface data - handle both v3 and v4 formats
	var iFaceStream []struct {
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
	}
	var iFaceCheck []struct {
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
	}
	if monitorIFaceShowResp.Results != nil {
		iFaceStream = monitorIFaceShowResp.Results.IFaceStream
		iFaceCheck = monitorIFaceShowResp.Results.IFaceCheck
	} else {
		iFaceStream = monitorIFaceShowResp.Data.IFaceStream
		iFaceCheck = monitorIFaceShowResp.Data.IFaceCheck
	}

	// Get lan devices data - handle both v3 and v4 formats
	var lanDevices []struct {
		ApName       string `json:"apname"`
		AcGid        int    `json:"ac_gid"`
		Mac          string `json:"mac"`
		LinkAddr     string `json:"link_addr"`
		Hostname     string `json:"hostname"`
		DTalkName    string `json:"dtalk_name"`
		DownRate     string `json:"downrate"`
		Reject       int    `json:"reject"`
		Uprate       string `json:"uprate"`
		Signal       string `json:"signal"`
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
	}
	if monitorLanIPShowResp.Results != nil && monitorLanIPShowResp.Results.Data != nil {
		lanDevices = monitorLanIPShowResp.Results.Data
	} else {
		lanDevices = monitorLanIPShowResp.Data.Data
	}

	ch <- prometheus.MustNewConstMetric(
		m.version,
		prometheus.GaugeValue,
		1,
		sysStat.VerInfo.Version, sysStat.VerInfo.Arch, sysStat.VerInfo.VerString,
	)

	{
		ch <- prometheus.MustNewConstMetric(
			m.up,
			prometheus.GaugeValue,
			1,
			"host",
		)
		ch <- prometheus.MustNewConstMetric(
			m.uptime,
			prometheus.GaugeValue,
			float64(sysStat.Uptime),
			"host",
		)
		ch <- prometheus.MustNewConstMetric(
			m.networkUploadTotalBytes,
			prometheus.GaugeValue,
			float64(sysStat.Stream.TotalUp),
			"host", "host", "host",
		)
		ch <- prometheus.MustNewConstMetric(
			m.networkDownloadTotalBytes,
			prometheus.GaugeValue,
			float64(sysStat.Stream.TotalDown),
			"host", "host", "host",
		)
		ch <- prometheus.MustNewConstMetric(
			m.networkUploadSpeedBytes,
			prometheus.GaugeValue,
			float64(sysStat.Stream.Upload),
			"host", "host", "host",
		)
		ch <- prometheus.MustNewConstMetric(
			m.networkDownloadSpeedBytes,
			prometheus.GaugeValue,
			float64(sysStat.Stream.Download),
			"host", "host", "host",
		)
		ch <- prometheus.MustNewConstMetric(
			m.networkConnectCount,
			prometheus.GaugeValue,
			float64(sysStat.Stream.ConnectNum),
			"host", "host", "host",
		)
	}

	if len(sysStat.Cpu) > 1 {
		sysStat.Cpu = sysStat.Cpu[1:]
	}
	for k, v := range sysStat.Cpu {
		s := v[:len(v)-1]
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			logger("cpuUsageRatio", "error: %s", err)
		}
		ch <- prometheus.MustNewConstMetric(
			m.cpuUsageRatio,
			prometheus.GaugeValue,
			f/100,
			fmt.Sprintf("core/%v", k),
		)
	}

	cpuTemp := 0.0
	if len(sysStat.CpuTemp) > 0 {
		cpuTemp = float64(sysStat.CpuTemp[0])
	}
	ch <- prometheus.MustNewConstMetric(
		m.cpuTemperature,
		prometheus.GaugeValue,
		cpuTemp,
	)

	ch <- prometheus.MustNewConstMetric(
		m.memorySizeKiloBytes,
		prometheus.GaugeValue,
		float64(sysStat.Memory.Total),
	)

	ch <- prometheus.MustNewConstMetric(
		m.memoryUsageKiloBytes,
		prometheus.GaugeValue,
		float64(sysStat.Memory.Total-sysStat.Memory.Available),
	)

	ch <- prometheus.MustNewConstMetric(
		m.memoryCachedKiloBytes,
		prometheus.GaugeValue,
		float64(sysStat.Memory.Cached),
	)

	ch <- prometheus.MustNewConstMetric(
		m.memoryBuffersKiloBytes,
		prometheus.GaugeValue,
		float64(sysStat.Memory.Buffers),
	)

	for _, i := range iFaceStream {
		internet := ""
		parentInterface := ""
		interfaceUp := 1
		interfaceID := fmt.Sprintf("interface/%s", i.Interface)
		interfaceUptime := int64(0)
		display := displayName(i.Interface)

		for _, n := range iFaceCheck {
			if n.Interface == i.Interface {
				internet = n.Internet
				parentInterface = n.ParentInterface
				if n.Result != "success" {
					interfaceUp = 0
				} else {
					if updateTime, err := strconv.Atoi(n.UpdateTime); err == nil {
						interfaceUptime = time.Now().Unix() - int64(updateTime)
					}
				}
			}
		}

		ch <- prometheus.MustNewConstMetric(
			m.interfaceInfo,
			prometheus.GaugeValue,
			1,
			interfaceID, i.Interface, i.Comment, internet, parentInterface, i.IpAddr, display,
		)

		ch <- prometheus.MustNewConstMetric(
			m.up,
			prometheus.GaugeValue,
			float64(interfaceUp),
			interfaceID,
		)

		ch <- prometheus.MustNewConstMetric(
			m.uptime,
			prometheus.GaugeValue,
			float64(interfaceUptime),
			interfaceID,
		)

		ch <- prometheus.MustNewConstMetric(
			m.networkUploadTotalBytes,
			prometheus.GaugeValue,
			float64(i.TotalUp),
			interfaceID, display, i.IpAddr,
		)

		ch <- prometheus.MustNewConstMetric(
			m.networkDownloadTotalBytes,
			prometheus.GaugeValue,
			float64(i.TotalDown),
			interfaceID, display, i.IpAddr,
		)

		ch <- prometheus.MustNewConstMetric(
			m.networkUploadSpeedBytes,
			prometheus.GaugeValue,
			float64(i.Upload),
			interfaceID, display, i.IpAddr,
		)

		ch <- prometheus.MustNewConstMetric(
			m.networkDownloadSpeedBytes,
			prometheus.GaugeValue,
			float64(i.Download),
			interfaceID, display, i.IpAddr,
		)

		if interfaceConnectCount, err := strconv.Atoi(i.ConnectNum); err == nil {
			ch <- prometheus.MustNewConstMetric(
				m.networkConnectCount,
				prometheus.GaugeValue,
				float64(interfaceConnectCount),
				interfaceID, display, i.IpAddr,
			)
		}
	}

	ch <- prometheus.MustNewConstMetric(
		m.deviceCount,
		prometheus.GaugeValue,
		float64(sysStat.OnlineUser.Count),
	)

	for _, i := range lanDevices {
		deviceID := fmt.Sprintf("device/%s", i.IpAddr)
		display := displayName(i.Comment, i.Hostname, i.IpAddr, i.Mac)

		ch <- prometheus.MustNewConstMetric(
			m.deviceInfo,
			prometheus.GaugeValue,
			1,
			deviceID, i.Mac, i.Hostname, i.IpAddr, i.Comment, display,
		)

		ch <- prometheus.MustNewConstMetric(
			m.networkUploadTotalBytes,
			prometheus.GaugeValue,
			float64(i.TotalUp),
			deviceID, display, i.IpAddr,
		)

		ch <- prometheus.MustNewConstMetric(
			m.networkUploadSpeedBytes,
			prometheus.GaugeValue,
			float64(i.Upload),
			deviceID, display, i.IpAddr,
		)

		ch <- prometheus.MustNewConstMetric(
			m.networkDownloadTotalBytes,
			prometheus.GaugeValue,
			float64(i.TotalDown),
			deviceID, display, i.IpAddr,
		)

		ch <- prometheus.MustNewConstMetric(
			m.networkDownloadSpeedBytes,
			prometheus.GaugeValue,
			float64(i.Download),
			deviceID, display, i.IpAddr,
		)

		ch <- prometheus.MustNewConstMetric(
			m.networkConnectCount,
			prometheus.GaugeValue,
			float64(i.ConnectNum),
			deviceID, display, i.IpAddr,
		)
	}
}

func newDesc(namespace string, metricName string, help string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(namespace+"_"+metricName, help, labels, nil)
}

func displayName(args ...string) string {
	for _, i := range args {
		if len(i) > 0 {
			return i
		}
	}
	return "unknown"
}
