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
	if mod.Result != 30000 {
		return nil, errors.New(mod.ErrMsg)
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
	if modV6.Result != 30000 {
		return nil, errors.New(modV6.ErrMsg)
	}

	mod.Data.Data = append(mod.Data.Data, modV6.Data.Data...)

	return &mod, nil
}
