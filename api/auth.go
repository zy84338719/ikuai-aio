package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
)

func (c *Client) Login() error {
	req := &LoginReq{
		Username: c.iKuaiUsername,
		Password: toMD5(c.iKuaiPassword),
		Pass:     base64.StdEncoding.EncodeToString([]byte("salt_11" + c.iKuaiPassword)),
	}
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	resp, err := c.request(iKuaiLoginPath, b)
	if err != nil {
		return err
	}

	var mod LoginResp
	if err = json.Unmarshal(resp, &mod); err != nil {
		return err
	}

	// Auto-detect version based on response format
	if mod.IsV4() {
		c.version = VersionV4
		log.Println("Detected iKuai OS version: v4")
	} else {
		c.version = VersionV3
		log.Println("Detected iKuai OS version: v3")
	}

	// Check for login success using IsSuccess()
	if !mod.IsSuccess() {
		return errors.New(mod.GetErrMsg())
	}

	return nil
}

func (c *Client) IsLogin() bool {
	r, err := c.WebUserShow()
	if err != nil {
		return false
	}

	// Check if login failed using IsLoginFailed()
	if r.IsLoginFailed() {
		return false
	}

	return true
}
