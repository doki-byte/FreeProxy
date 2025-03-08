package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

type Config struct {
	Email     string
	FofaKey   string
	HunterKey string
	QuakeKey  string
	Country   string
	Maxpage   string

	CoroutineCount int
	LiveProxies    int
	AllProxies     int
	LiveProxyLists []string
	Timeout        string
	SocksAddress   string
	FilePath       string

	Status int

	Code        int
	Error       string
	GlobalProxy string
}

// 获取当前执行程序所在的绝对路径
func GetCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

func CreateConfigFile() error {
	// 获取配置文件路径
	optSys := runtime.GOOS
	path := ""
	if optSys == "windows" {
		path = GetCurrentAbPathByExecutable() + "\\config.ini"
	} else {
		path = GetCurrentAbPathByExecutable() + "/config.ini"
	}

	// 检查文件是否已存在
	if _, err := os.Stat(path); err == nil {
		// 如果文件存在，则返回
		return nil
	}

	// 文件不存在，创建并写入默认配置
	defaultConfig := map[string]string{
		"Timeout":        "10",
		"GlobalProxy":    "0",
		"Country":        "0",
		"Email":          "",
		"FofaKey":        "",
		"HunterKey":      "",
		"QuakeKey":       "",
		"Maxpage":        "10",
		"CoroutineCount": "200",
		"SocksAddress":   "socks5://127.0.0.1:1080",
	}

	// 创建文件
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// 写入默认配置
	for key, value := range defaultConfig {
		_, err := writer.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if err != nil {
			return fmt.Errorf("failed to write to config file: %v", err)
		}
	}
	writer.Flush()

	return nil
}

func GetConfig() *Config {
	// 创建一个空的结构体
	c := &Config{}
	cr := reflect.ValueOf(c).Elem()

	// 读取 config.ini 文件路径
	optSys := runtime.GOOS
	path := ""
	if optSys == "windows" {
		path = GetCurrentAbPathByExecutable() + "\\config.ini"
	} else {
		path = GetCurrentAbPathByExecutable() + "/config.ini"
	}

	// 打印路径，检查是否正确
	log.Printf("Config file path: %s", path)

	// 打开配置文件
	f, err := os.Open(path)
	if err != nil {
		// 如果文件不存在，创建文件
		log.Printf("Config file not found, attempting to create it... Error: %v", err)
		err := CreateConfigFile()
		if err != nil {
			log.Fatalf("Failed to create config file: %v", err)
			return nil
		}
		// 尝试重新打开文件
		f, err = os.Open(path)
		if err != nil {
			log.Fatalf("Failed to open config file after creation: %v", err)
			return nil
		}
	}
	defer f.Close()

	// 逐行读取文件内容
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		index := strings.Index(line, "=")
		if index == -1 {
			continue // 跳过不符合格式的行
		}

		key := strings.TrimSpace(line[:index])
		value := strings.TrimSpace(line[index+1:])

		// 获取结构体字段
		field := cr.FieldByName(key)
		if !field.IsValid() { // 检查字段是否存在
			log.Printf("Warning: Unknown config key: %s", key)
			continue
		}

		// 设置字段值
		if field.Kind() == reflect.String {
			field.SetString(value)
		} else if field.Kind() == reflect.Int {
			intValue, err := strconv.Atoi(value)
			if err != nil {
				log.Printf("Warning: Invalid integer value for key %s: %s", key, value)
				continue
			}
			field.SetInt(int64(intValue))
		} else if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.String {
			field.Set(reflect.ValueOf(strings.Split(value, ",")))
		}
	}

	if err = s.Err(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	return c
}

func (p *Config) SaveConfig() error {
	optSys := runtime.GOOS
	path := ""
	if optSys == "windows" {
		path = GetCurrentAbPathByExecutable() + "\\config.ini"
	} else {
		path = GetCurrentAbPathByExecutable() + "/config.ini"
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	configMap := map[string]string{
		"Email":          p.Email,
		"FofaKey":        p.FofaKey,
		"HunterKey":      p.HunterKey,
		"QuakeKey":       p.QuakeKey,
		"Maxpage":        p.Maxpage,
		"CoroutineCount": strconv.Itoa(p.CoroutineCount),
		"Timeout":        p.Timeout,
		"SocksAddress":   p.SocksAddress,
		"Status":         strconv.Itoa(p.Status),
		"Code":           strconv.Itoa(p.Code),
		"Error":          p.Error,
		"Country":        p.Country,
		"GlobalProxy":    p.GlobalProxy,
	}

	for key, value := range configMap {
		//if key == "SocksAddress" {
		//	value = strings.Replace(value, "socks5://", "", -1)
		//}

		_, err := writer.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if err != nil {
			return err
		}
	}
	writer.Flush()
	return nil
}

func (p *Config) GetProfile() Config {
	//optSys := runtime.GOOS
	//proxy_success_path := ""
	//if optSys == "windows" {
	//	proxy_success_path = GetCurrentAbPathByExecutable() + "\\proxy_success.txt"
	//} else {
	//	proxy_success_path = GetCurrentAbPathByExecutable() + "/proxy_success.txt"
	//}
	//if _, err := os.Stat(proxy_success_path); err == nil {
	//	// 如果文件存在，则返回
	//	p.FilePath = proxy_success_path
	//} else {
	//	p.FilePath = ""
	//}
	return *p
}

func (p *Config) GetCoroutineCount() int {
	return p.CoroutineCount
}

func (p *Config) GetLiveProxies() int {
	return p.LiveProxies
}

func (p *Config) SetAllProxies(datasets []string) {
	p.AllProxies = len(datasets)
	p.LiveProxyLists = datasets
}

func (p *Config) SetLiveProxies(datasets []string) {
	p.LiveProxyLists = datasets
	p.LiveProxies = len(datasets)
}

func (p *Config) GetTimeout() string { return p.Timeout }

func (p *Config) GetSocksAddress() string {
	return p.SocksAddress
}

func (p *Config) GetStatus() int {
	return p.Status
}

func (p *Config) SetStatus(i int) {
	p.Status = i
}
