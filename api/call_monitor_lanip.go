package api

import (
	"encoding/json"
	"errors"
)

func (c *Client) MonitorLanIPShow() (*MonitorLanIPShowResp, error) {
	req := &CallReq{
		FuncName: "monitor_lanip",
		Action:   "show",
		Param: map[string]string{
			"TYPE": "data",
		},
	}
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	resp, err := c.request(iKuaiCallPath, b)
	if err != nil {
		return nil, err
	}

	var mod MonitorLanIPShowResp
	if err = json.Unmarshal(resp, &mod); err != nil {
		return nil, err
	}
	// Check for success using IsSuccess()
	if !mod.IsSuccess() {
		return nil, errors.New(mod.GetErrMsg())
	}

	reqV6 := &CallReq{
		FuncName: "monitor_lanipv6",
		Action:   "show",
		Param: map[string]string{
			"TYPE": "data",
		},
	}
	bV6, err := json.Marshal(reqV6)
	if err != nil {
		return nil, err
	}
	respV6, err := c.request(iKuaiCallPath, bV6)
	if err != nil {
		return nil, err
	}
	var modV6 MonitorLanIPShowResp
	if err = json.Unmarshal(respV6, &modV6); err != nil {
		return nil, err
	}
	// Check for success using IsSuccess()
	if !modV6.IsSuccess() {
		return nil, errors.New(modV6.GetErrMsg())
	}

	// Handle both v3 and v4 formats when combining IPv4 and IPv6 data
	if mod.Results != nil && mod.Results.Data != nil {
		if modV6.Results != nil && modV6.Results.Data != nil {
			mod.Results.Data = append(mod.Results.Data, modV6.Results.Data...)
		} else if modV6.Data.Data != nil {
			// v6 is v3 format, convert to v4
			for _, item := range modV6.Data.Data {
				mod.Results.Data = append(mod.Results.Data, item)
			}
		}
	} else if mod.Data.Data != nil {
		if modV6.Data.Data != nil {
			mod.Data.Data = append(mod.Data.Data, modV6.Data.Data...)
		} else if modV6.Results != nil && modV6.Results.Data != nil {
			// v6 is v4 format, convert to v3
			for _, item := range modV6.Results.Data {
				mod.Data.Data = append(mod.Data.Data, item)
			}
		}
	}

	return &mod, nil
}
