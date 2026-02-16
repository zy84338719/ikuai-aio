package job

import (
	"time"

	"github.com/NERVEbing/ikuai-aio/api"
	"github.com/NERVEbing/ikuai-aio/config"
)

// CustomISPItem represents a custom ISP item (same as in api package)
type CustomISPItem struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	IPGroup string `json:"ipgroup"`
	Comment string `json:"comment"`
	Time    string `json:"time"`
}

func updateCustomISP(c *config.IKuaiCronCustomISP, tag string) error {
	var rows []string
	start := time.Now()
	for _, url := range c.Url {
		r, err := fetch(url)
		if err != nil {
			logger(tag, "fetch %s failed, error: %s", url, err)
			continue
		}
		logger(tag, "fetch %s success, rows: %d", url, len(r))
		rows = append(rows, r...)
	}
	logger(tag, "fetch total rows: %d", len(rows))
	if len(rows) == 0 {
		return nil
	}

	client := api.NewClient()
	if err := client.Login(); err != nil {
		return err
	}
	customISPShowResp, err := client.CustomISPShow()
	if err != nil {
		return err
	}
	var ids []int
	for _, i := range customISPShowResp.GetData() {
		if v, ok := i.(CustomISPItem); ok && v.Name == c.Name {
			ids = append(ids, v.ID)
		}
	}
	if err = client.CustomISPDel(ids); err != nil {
		return err
	}
	count, err := client.CustomISPAdd(c.Name, rows, c.Comment)
	if err != nil {
		return err
	}
	logger(tag, "add custom isp unique rows count: %d, duration: %s", count, time.Now().Sub(start).String())

	return nil
}
