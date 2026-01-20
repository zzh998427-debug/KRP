package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

// Config模板结构体
type Config struct {
	UUID        string
	Port        string
	FakeDomain  string
	ServerNames string
	Fingerprint string
	PublicKey   string
	ShortID     string
	Protocol    string // reality 或 ws
	CFCDN       string
}

func main() {
	// 从环境变量读取（默认随机生成）
	uuid := getEnvOrRandom("UUID", 36) // UUID格式
	port := getEnv("PORT", "443")
	fakeDomain := getEnv("FAKE_DOMAIN", "www.microsoft.com")
	serverNames := getEnv("SERVER_NAMES", "www.microsoft.com,www.google.com")
	fingerprint := getEnv("FINGERPRINT", "chrome")
	publicKey := getEnvOrRandom("PUBLIC_KEY", 64) // Reality公钥
	shortID := getEnvOrRandom("SHORT_ID", 16)
	protocol := getEnv("FALLBACK_PROTO", "") // 空为reality
	cfCDN := getEnv("CF_CDN", "")

	// 模块化配置
	conf := Config{
		UUID:        uuid,
		Port:        port,
		FakeDomain:  fakeDomain,
		ServerNames: serverNames,
		Fingerprint: fingerprint,
		PublicKey:   publicKey,
		ShortID:     shortID,
		Protocol:    protocol,
		CFCDN:       cfCDN,
	}

	// 生成config.json
	tmpl, err := template.ParseFiles("/config.json.template")
	if err != nil {
		log.Fatal("Template parse error:", err)
	}
	f, err := os.Create("/config.json")
	if err != nil {
		log.Fatal("Config create error:", err)
	}
	defer f.Close()
	err = tmpl.Execute(f, conf)
	if err != nil {
		log.Fatal("Template execute error:", err)
	}

	// 启动Xray（无日志到盘，最小输出）
	cmd := exec.Command("/usr/bin/xray", "-config", "/config.json")
	cmd.Stdout = os.Stdout // 输出到Koyeb logs
	cmd.Stderr = os.Stderr
	go func() { // 断线重连：Xray内置，但这里监控退出
		if err := cmd.Run(); err != nil {
			log.Println("Xray error:", err)
		}
	}()

	// 输出节点链接（在logs中打印）
	domain := os.Getenv("KOYEB_PUBLIC_DOMAIN") // Koyeb环境变量
	if domain == "" {
		domain = "your-koyeb-domain.koyeb.app" // 占位
	}
	link := fmt.Sprintf("vless://%s@%s:%s?security=reality&fp=%s&pbk=%s&sni=%s&sid=%s#Koyeb-Node",
		uuid, domain, port, fingerprint, publicKey, fakeDomain, shortID)
	fmt.Println("Node Link:", link)

	// 小型Web服务器（可选，轻量输出链接后关闭）
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, link) })
	// go http.ListenAndServe(":8080", nil) // 如果需要Web路径

	// 等待（保持运行）
	select {}
}

// 辅助函数：环境变量或随机
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvOrRandom(key string, length int) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	b := make([]byte, length/2)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("Random error:", err)
	}
	return hex.EncodeToString(b)
}