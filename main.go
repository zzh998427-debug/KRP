package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type Config struct {
	UUID         string
	Port         string
	FakeDomain   string
	ServerNames  string
	Fingerprint  string
	PublicKey    string
	PrivateKey   string // 新增：对应私钥
	ShortID      string
	Protocol     string // ws 或空
}

func main() {
	// 环境变量读取
	uuid := getEnvOrRandom("UUID", 36)
	port := getEnv("PORT", "443")
	fakeDomain := getEnv("FAKE_DOMAIN", "www.microsoft.com")
	serverNames := getEnv("SERVER_NAMES", "www.microsoft.com,www.google.com")
	fingerprint := getEnv("FINGERPRINT", "chrome")

	// Reality 密钥对生成（如果不设）
	publicKey, privateKey := getRealityKeys()
	if pk := getEnv("PUBLIC_KEY", ""); pk != "" {
		publicKey = pk
	}
	if sk := getEnv("PRIVATE_KEY", ""); sk != "" {
		privateKey = sk
	}

	shortID := getEnvOrRandom("SHORT_ID", 8) // 缩短为8位，够用
	protocol := getEnv("FALLBACK_PROTO", "ws") // Fly强制ws

	conf := Config{
		UUID:        uuid,
		Port:        port,
		FakeDomain:  fakeDomain,
		ServerNames: serverNames,
		Fingerprint: fingerprint,
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
		ShortID:     shortID,
		Protocol:    protocol,
	}

	// 生成 config.json
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

	// 启动 Xray
	cmd := exec.Command("/usr/bin/xray", "run", "-config", "/config.json")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatal("Xray start error:", err)
	}

	// 输出节点信息
	domain := getEnv("DOMAIN", "")
	if domain == "" {
		log.Println("WARNING: 请在 Fly Dashboard Secrets 设置 DOMAIN=your-app.fly.dev")
		domain = "your-app.fly.dev" // 占位
	}

	path := "/ws" + shortID

	var link string
	if protocol == "ws" {
		link = fmt.Sprintf("vless://%s@%s:443?type=ws&security=none&path=%s&host=%s#Fly-WS-Node",
			uuid, domain, path, fakeDomain)
	} else {
		// Reality 备用（Fly 不推荐）
		link = fmt.Sprintf("vless://%s@%s:443?security=reality&fp=%s&pbk=%s&sni=%s&sid=%s#Fly-Reality-Node",
			uuid, domain, fingerprint, publicKey, fakeDomain, shortID)
	}

	fmt.Println("======================================")
	fmt.Println("Domain:", domain)
	fmt.Println("WS Path:", path)
	fmt.Println("Node Link:", link)
	fmt.Println("======================================")

	// 等待 Xray 进程
	if err := cmd.Wait(); err != nil {
		log.Println("Xray exited:", err)
	}
}

// 辅助函数
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
	if _, err := rand.Read(b); err != nil {
		log.Fatal("Random error:", err)
	}
	return hex.EncodeToString(b)
}

// 生成 Reality ed25519 密钥对
func getRealityKeys() (public string, private string) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal("Key generate error:", err)
	}
	return hex.EncodeToString(pub), hex.EncodeToString(priv)
}