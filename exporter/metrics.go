package exporter

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/NERVEbing/ikuai-aio/api"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	client *api.Client

	version       *prometheus.Desc
	up            *prometheus.Desc
	uptime        *prometheus.Desc

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

func NewMetrics(namespace string, client *api.Client) *Metrics {
	return &Metrics{
		client: client,
		version:                   prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "version"), "Router version info", []string{"version", "arch", "ver_string"}, nil),
		up:                        prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "up"), "Router up status", []string{"id"}, nil),
		uptime:                    prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "uptime"), "Router uptime in seconds", []string{"id"}, nil),
		cpuUsageRatio:             prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "cpu_usage_ratio"), "CPU usage ratio", []string{"id"}, nil),
		cpuTemperature:            prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "cpu_temperature"), "CPU temperature", nil, nil),
		memorySizeKiloBytes:       prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "memory_size_kilo_bytes"), "Router memory size in KB", nil, nil),
		memoryUsageKiloBytes:      prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "memory_usage_kilo_bytes"), "Router memory used in KB", nil, nil),
		memoryCachedKiloBytes:     prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "memory_cached_kilo_bytes"), "Router memory cached in KB", nil, nil),
		memoryBuffersKiloBytes:    prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "memory_buffers_kilo_bytes"), "Router memory buffers in KB", nil, nil),
		interfaceInfo:             prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "interface_info"), "Network interface info", []string{"id", "interface", "comment", "internet", "parent_interface", "ip_addr", "display"}, nil),
		deviceCount:               prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "device_count"), "Total number of devices", nil, nil),
		deviceInfo:                prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "device_info"), "LAN device info", []string{"id", "mac", "hostname", "ip_addr", "comment", "display"}, nil),
		networkUploadTotalBytes:   prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "network_upload_total_bytes"), "Total network upload in bytes", []string{"id", "display", "ip_addr"}, nil),
		networkDownloadTotalBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "network_download_total_bytes"), "Total network download in bytes", []string{"id", "display", "ip_addr"}, nil),
		networkUploadSpeedBytes:   prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "network_upload_speed_bytes"), "Current upload speed in bytes/s", []string{"id", "display", "ip_addr"}, nil),
		networkDownloadSpeedBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "network_download_speed_bytes"), "Current download speed in bytes/s", []string{"id", "display", "ip_addr"}, nil),
		networkConnectCount:       prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "network_connect_count"), "Network connection count", []string{"id", "display", "ip_addr"}, nil),
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
			log.Printf("Recovered from panic in metrics collection: %v", err)
			ch <- prometheus.MustNewConstMetric(
				m.up,
				prometheus.GaugeValue,
				0,
				"host",
			)
		}
	}()

	if !m.client.IsLogin() {
		log.Println("Cookie has expired, attempting to log in again")
		if err := m.client.Login(); err != nil {
			log.Printf("Login failed: %v", err)
			ch <- prometheus.MustNewConstMetric(
				m.up,
				prometheus.GaugeValue,
				0,
				"host",
			)
			return
		}
		log.Println("Login successful")
	}

	homepageShowSysStatResp, err := m.client.HomepageShowSysStat()
	if err != nil {
		log.Printf("Error getting homepage sysstat: %v", err)
		return
	}

	monitorLanIPShowResp, err := m.client.MonitorLanIPShow()
	if err != nil {
		log.Printf("Error getting LAN devices: %v", err)
		monitorLanIPShowResp = nil
	}

	monitorIFaceShowResp, err := m.client.MonitorIFaceShow()
	if err != nil {
		log.Printf("Error getting network interfaces: %v", err)
		monitorIFaceShowResp = nil
	}

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

	data := homepageShowSysStatResp.GetData()
	dataBytes, err := json.Marshal(data)
	if err == nil {
		var dataMap map[string]interface{}
		if err := json.Unmarshal(dataBytes, &dataMap); err == nil {
			if sysStatMap, ok := dataMap["sysstat"].(map[string]interface{}); ok {
				sysStat.Cpu = getJSONStringSlice(sysStatMap, "cpu")
				sysStat.CpuTemp = getJSONIntSlice(sysStatMap, "cputemp")
				sysStat.Freq = getJSONStringSlice(sysStatMap, "freq")
				sysStat.GWid = getJSONString(sysStatMap, "gwid")
				sysStat.Hostname = getJSONString(sysStatMap, "hostname")
				sysStat.LinkStatus = getJSONInt(sysStatMap, "link_status")

				if memMap, ok := sysStatMap["memory"].(map[string]interface{}); ok {
					sysStat.Memory.Total = getJSONInt64(memMap, "total")
					sysStat.Memory.Available = getJSONInt64(memMap, "available")
					sysStat.Memory.Free = getJSONInt64(memMap, "free")
					sysStat.Memory.Cached = getJSONInt64(memMap, "cached")
					sysStat.Memory.Buffers = getJSONInt64(memMap, "buffers")
					sysStat.Memory.Used = getJSONString(memMap, "used")
				}

				if onlineMap, ok := sysStatMap["online_user"].(map[string]interface{}); ok {
					sysStat.OnlineUser.Count = getJSONInt(onlineMap, "count")
					sysStat.OnlineUser.Count2G = getJSONInt(onlineMap, "count_2g")
					sysStat.OnlineUser.Count5G = getJSONInt(onlineMap, "count_5g")
					sysStat.OnlineUser.CountWired = getJSONInt(onlineMap, "count_wired")
					sysStat.OnlineUser.CountWireless = getJSONInt(onlineMap, "count_wireless")
				}

				if streamMap, ok := sysStatMap["stream"].(map[string]interface{}); ok {
					sysStat.Stream.ConnectNum = getJSONInt(streamMap, "connect_num")
					sysStat.Stream.Upload = getJSONInt(streamMap, "upload")
					sysStat.Stream.Download = getJSONInt(streamMap, "download")
					sysStat.Stream.TotalUp = getJSONInt64(streamMap, "total_up")
					sysStat.Stream.TotalDown = getJSONInt64(streamMap, "total_down")
				}

				sysStat.Uptime = getJSONInt(sysStatMap, "uptime")

				if verMap, ok := sysStatMap["verinfo"].(map[string]interface{}); ok {
					sysStat.VerInfo.ModelName = getJSONString(verMap, "modelname")
					sysStat.VerInfo.VerString = getJSONString(verMap, "verstring")
					sysStat.VerInfo.Version = getJSONString(verMap, "version")
					sysStat.VerInfo.BuildDate = getJSONInt64(verMap, "build_date")
					sysStat.VerInfo.Arch = getJSONString(verMap, "arch")
					sysStat.VerInfo.SysBit = getJSONString(verMap, "sysbit")
					sysStat.VerInfo.VerFlags = getJSONString(verMap, "verflags")
					sysStat.VerInfo.IsEnterprise = getJSONInt(verMap, "is_enterprise")
					sysStat.VerInfo.SupportI18N = getJSONInt(verMap, "support_i18n")
					sysStat.VerInfo.SupportLcd = getJSONInt(verMap, "support_lcd")
				}
			}
		}
	}

	ch <- prometheus.MustNewConstMetric(
		m.up,
		prometheus.GaugeValue,
		1,
		"host",
	)

	if sysStat.VerInfo.VerString != "" {
		ch <- prometheus.MustNewConstMetric(
			m.version,
			prometheus.GaugeValue,
			1,
			sysStat.VerInfo.Version,
			sysStat.VerInfo.Arch,
			sysStat.VerInfo.VerString,
		)
	}

	if sysStat.Uptime > 0 {
		ch <- prometheus.MustNewConstMetric(
			m.uptime,
			prometheus.GaugeValue,
			float64(sysStat.Uptime),
			"host",
		)
		log.Printf("Uptime: %d seconds", sysStat.Uptime)
	}

	// CPU metrics - export each core separately
	for i, cpu := range sysStat.Cpu {
		cpuValue, err := parseCPUUsage(cpu)
		if err == nil {
			ch <- prometheus.MustNewConstMetric(
				m.cpuUsageRatio,
				prometheus.GaugeValue,
				cpuValue,
				fmt.Sprintf("core/%d", i),
			)
		}
	}

	if len(sysStat.CpuTemp) > 0 {
		ch <- prometheus.MustNewConstMetric(
			m.cpuTemperature,
			prometheus.GaugeValue,
			float64(sysStat.CpuTemp[0]),
		)
		log.Printf("CPU temperature: %d°C", sysStat.CpuTemp[0])
	}

	// Memory metrics
	if sysStat.Memory.Total > 0 {
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
		log.Printf("Memory: Total: %d KB, Used: %d KB, Cached: %d KB, Available: %d KB",
			sysStat.Memory.Total/1024,
			(sysStat.Memory.Total-sysStat.Memory.Available)/1024,
			sysStat.Memory.Cached/1024,
			sysStat.Memory.Available/1024,
		)
	}

	// Collect interface metrics
	if monitorIFaceShowResp != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic processing interfaces: %v", r)
				}
			}()
			ifaceData := monitorIFaceShowResp.GetData()
			ifaceDataBytes, err := json.Marshal(ifaceData)
			if err == nil {
				var dataMap map[string]interface{}
				if err := json.Unmarshal(ifaceDataBytes, &dataMap); err == nil {
					if ifaceCheck, ok := dataMap["iface_check"].([]interface{}); ok {
						for _, iface := range ifaceCheck {
							if ifaceMap, ok := iface.(map[string]interface{}); ok {
								id := getJSONString(ifaceMap, "id")
								interfaceName := getJSONString(ifaceMap, "interface")
								comment := getJSONString(ifaceMap, "comment")
								internet := getJSONString(ifaceMap, "internet")
								parentInterface := getJSONString(ifaceMap, "parent_interface")
								ipAddr := getJSONString(ifaceMap, "ip_addr")
								display := getJSONString(ifaceMap, "comment")
								if display == "" {
									display = interfaceName
								}

								ch <- prometheus.MustNewConstMetric(
									m.interfaceInfo,
									prometheus.GaugeValue,
									1,
									id, interfaceName, comment, internet, parentInterface, ipAddr, display,
								)
							}
						}
					}

					// Export network metrics for each interface
					if ifaceStream, ok := dataMap["iface_stream"].([]interface{}); ok {
						for _, iface := range ifaceStream {
							if ifaceMap, ok := iface.(map[string]interface{}); ok {
								interfaceName := getJSONString(ifaceMap, "interface")
								ipAddr := getJSONString(ifaceMap, "ip_addr")
								display := getJSONString(ifaceMap, "comment")
								if display == "" {
									display = interfaceName
								}

								uploadTotal := getJSONInt64(ifaceMap, "total_up")
								downloadTotal := getJSONInt64(ifaceMap, "total_down")
								uploadSpeed := getJSONInt(ifaceMap, "upload")
								downloadSpeed := getJSONInt(ifaceMap, "download")
								connectNum := getJSONInt(ifaceMap, "connect_num")

								id := fmt.Sprintf("interface/%s", interfaceName)

								// Export up metric for interface
								ch <- prometheus.MustNewConstMetric(
									m.up,
									prometheus.GaugeValue,
									1,
									id,
								)

								// Export uptime metric for interface (set to 0 since API doesn't provide it)
								ch <- prometheus.MustNewConstMetric(
									m.uptime,
									prometheus.GaugeValue,
									0,
									id,
								)

								ch <- prometheus.MustNewConstMetric(
									m.networkUploadTotalBytes,
									prometheus.CounterValue,
									float64(uploadTotal),
									id, display, ipAddr,
								)
								ch <- prometheus.MustNewConstMetric(
									m.networkDownloadTotalBytes,
									prometheus.CounterValue,
									float64(downloadTotal),
									id, display, ipAddr,
								)
								ch <- prometheus.MustNewConstMetric(
									m.networkUploadSpeedBytes,
									prometheus.GaugeValue,
									float64(uploadSpeed),
									id, display, ipAddr,
								)
								ch <- prometheus.MustNewConstMetric(
									m.networkDownloadSpeedBytes,
									prometheus.GaugeValue,
									float64(downloadSpeed),
									id, display, ipAddr,
								)
								ch <- prometheus.MustNewConstMetric(
									m.networkConnectCount,
									prometheus.GaugeValue,
									float64(connectNum),
									id, display, ipAddr,
								)
							}
						}
					}
				}
			}
		}()
	}

	// Collect LAN device metrics
	deviceCount := 0
	if monitorLanIPShowResp != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic processing LAN devices: %v", r)
				}
			}()
			lanData := monitorLanIPShowResp.GetData()
			deviceCount = len(lanData)
			log.Printf("Found %d LAN devices", deviceCount)
			for i, device := range lanData {
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("Panic processing device: %v", r)
						}
					}()

					// Convert device struct to map
					deviceBytes, err := json.Marshal(device)
					if err != nil {
						log.Printf("Failed to marshal device %d: %v", i, err)
						return
					}
					var deviceMap map[string]interface{}
					if err := json.Unmarshal(deviceBytes, &deviceMap); err != nil {
						log.Printf("Failed to unmarshal device %d: %v", i, err)
						return
					}

					id := getJSONString(deviceMap, "id")
					mac := getJSONString(deviceMap, "mac")
					hostname := getJSONString(deviceMap, "hostname")
					ipAddr := getJSONString(deviceMap, "ip_addr")
					comment := getJSONString(deviceMap, "comment")
					display := hostname
					if display == "" {
						display = ipAddr
					}

					// Prefix id with "device/" to match expected format
					deviceId := fmt.Sprintf("device/%s", id)

					ch <- prometheus.MustNewConstMetric(
						m.deviceInfo,
						prometheus.GaugeValue,
						1,
						deviceId, mac, hostname, ipAddr, comment, display,
					)

					// Export network metrics for each device
					display = ipAddr
					uploadTotal := getJSONInt64(deviceMap, "total_up")
					downloadTotal := getJSONInt64(deviceMap, "total_down")
					uploadSpeed := getJSONInt(deviceMap, "upload")
					downloadSpeed := getJSONInt(deviceMap, "download")
					connectNum := getJSONInt(deviceMap, "connect_num")

					ch <- prometheus.MustNewConstMetric(
						m.networkUploadTotalBytes,
						prometheus.CounterValue,
						float64(uploadTotal),
						id, display, ipAddr,
					)
					ch <- prometheus.MustNewConstMetric(
						m.networkDownloadTotalBytes,
						prometheus.CounterValue,
						float64(downloadTotal),
						id, display, ipAddr,
					)
					ch <- prometheus.MustNewConstMetric(
						m.networkUploadSpeedBytes,
						prometheus.GaugeValue,
						float64(uploadSpeed),
						id, display, ipAddr,
					)
					ch <- prometheus.MustNewConstMetric(
						m.networkDownloadSpeedBytes,
						prometheus.GaugeValue,
						float64(downloadSpeed),
						id, display, ipAddr,
					)
					ch <- prometheus.MustNewConstMetric(
						m.networkConnectCount,
						prometheus.GaugeValue,
						float64(connectNum),
						id, display, ipAddr,
					)
				}()
			}
		}()
	}

	// Device count metric
	ch <- prometheus.MustNewConstMetric(
		m.deviceCount,
		prometheus.GaugeValue,
		float64(deviceCount),
	)

	log.Printf("Metrics collection completed. Devices: %d, Online Users: %d",
		deviceCount, sysStat.OnlineUser.Count)
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func getJSONString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
		if num, ok := val.(float64); ok {
			return strconv.FormatFloat(num, 'f', -1, 64)
		}
	}
	return ""
}

func getJSONInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

func getJSONInt64(m map[string]interface{}, key string) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		case int:
			return int64(v)
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return i
			}
		}
	}
	return 0
}

func getJSONStringSlice(m map[string]interface{}, key string) []string {
	if val, ok := m[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return []string{}
}

func getJSONIntSlice(m map[string]interface{}, key string) []int {
	if val, ok := m[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			result := make([]int, 0, len(slice))
			for _, item := range slice {
				switch v := item.(type) {
				case float64:
					result = append(result, int(v))
				case int:
					result = append(result, v)
				}
			}
			return result
		}
	}
	return []int{}
}

func parseCPUUsage(cpuStr string) (float64, error) {
	cpuStr = strings.TrimSpace(cpuStr)
	if len(cpuStr) > 0 && cpuStr[len(cpuStr)-1] == '%' {
		cpuStr = cpuStr[:len(cpuStr)-1]
	}
	var val float64
	_, err := fmt.Sscanf(cpuStr, "%f", &val)
	if err != nil {
		return 0, fmt.Errorf("invalid CPU format: %s", cpuStr)
	}
	return val, nil
}
