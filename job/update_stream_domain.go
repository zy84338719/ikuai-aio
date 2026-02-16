package job

import (
	"time"

	"github.com/NERVEbing/ikuai-aio/api"
	"github.com/NERVEbing/ikuai-aio/config"
)

// StreamDomainItem represents a stream domain item (same as in api package)
type StreamDomainItem struct {
	ID        int    `json:"id"`
	Interface string `json:"interface"`
	SrcAddr   string `json:"src_addr"`
	Enabled   string `json:"enabled"`
	Week      string `json:"week"`
	Comment   string `json:"comment"`
	Domain    string `json:"domain"`
	Time      string `json:"time"`
}

func updateStreamDomain(c *config.IKuaiCronStreamDomain, tag string) error {
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
	StreamDomainResp, err := client.StreamDomainShow()
	if err != nil {
		return err
	}
	var ids []int
	for _, i := range StreamDomainResp.GetData() {
		if v, ok := i.(StreamDomainItem); ok && v.Comment == c.Comment {
			ids = append(ids, v.ID)
		}
	}
	if err = client.StreamDomainDel(ids); err != nil {
		return err
	}
	count, err := client.StreamDomainAdd(c.Interface, rows, c.SrcAddr, c.Comment)
	if err != nil {
		return err
	}
	logger(tag, "add stream domain unique rows count: %d, duration: %s", count, time.Now().Sub(start).String())

	return nil
}
