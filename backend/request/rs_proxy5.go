/*
 * @Author: Lockly
 * @Date: 2025-02-17 18:56:25
 * @LastEditors: Lockly
 * @LastEditTime: 2025-02-17 18:57:39
 */

package request

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/imroc/req/v3"
)

type FreeProxy5Proxy struct {
	page int
	url  string
	reg  string
}

func (fp *FreeProxy5Proxy) Name() string {
	return "齐云代理"
}

var FreeProxy5 = FreeProxy5Proxy{
	page: 20,
	url:  "https://proxy.ip3366.net/free/?action=china&page=",
	reg:  `<td data-title="IP">([\d.]+)</td>\s*<td data-title="PORT">(\d+)</td>`,
}

func (fp *FreeProxy5Proxy) Fetch() ([]string, error) {
	var wg sync.WaitGroup
	ch := make(chan []string, fp.page)
	errCh := make(chan error, fp.page)

	for i := 1; i <= fp.page; i++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			proxies, err := fp.fetchPage(page)
			if err != nil {
				errCh <- err
				return
			}
			ch <- proxies
		}(i)
	}

	go func() {
		wg.Wait()
		close(ch)
		close(errCh)
	}()

	var allProxies []string
	for {
		select {
		case proxies, ok := <-ch:
			if !ok {
				return allProxies, nil
			}
			allProxies = append(allProxies, proxies...)
		case err, ok := <-errCh:
			if !ok {
				continue
			}
			return nil, err
		}
	}
}

func (fp *FreeProxy5Proxy) fetchPage(page int) ([]string, error) {
	client := req.C()
	resp, err := client.R().Get(fmt.Sprintf("%s%d", fp.url, page))
	if err != nil {
		return nil, err
	}

	body := resp.String()

	ipPortRegex := regexp.MustCompile(fp.reg)
	matches := ipPortRegex.FindAllStringSubmatch(body, -1)

	var proxies []string
	for _, match := range matches {
		if len(match) == 3 {
			ip := match[1]
			port := match[2]
			proxies = append(proxies, fmt.Sprintf("%s:%s", ip, port))
		}
	}

	return proxies, nil
}
