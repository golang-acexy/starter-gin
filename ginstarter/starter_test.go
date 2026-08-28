package ginstarter

import (
	"errors"
	"testing"
	"time"
)

func resetGinStarterForTest(t *testing.T) {
	t.Helper()
	if runtime := ginRuntimeState.Swap(nil); runtime != nil {
		_ = runtime.server.Close()
	}
}

func TestGinStarterLifecycle(t *testing.T) {
	resetGinStarterForTest(t)
	defer resetGinStarterForTest(t)

	starter := &GinStarter{Config: GinConfig{ListenAddress: "127.0.0.1:0"}}
	instance, err := starter.Start()
	if err != nil {
		t.Fatalf("启动 Gin Starter 失败：%v", err)
	}
	if RawGinEngine() != instance {
		t.Fatal("原始 Gin Engine 与运行时快照不一致")
	}
	if _, err = starter.Start(); !errors.Is(err, ErrGinStarterAlreadyStarted) {
		t.Fatalf("重复启动错误 = %v，期望 ErrGinStarterAlreadyStarted", err)
	}
	gracefully, stopped, err := starter.Stop(time.Second)
	if err != nil || !gracefully || !stopped {
		t.Fatalf("停止结果 = (%v, %v, %v)", gracefully, stopped, err)
	}
	if RawGinEngine() != nil {
		t.Fatal("停止后仍可读取 Gin Engine")
	}
}

func TestGinStarterConfigIsCachedCopy(t *testing.T) {
	starter := &GinStarter{Config: GinConfig{ListenAddress: "original"}}
	config := starter.getConfig()
	starter.Config.ListenAddress = "changed"
	if config.ListenAddress != "original" {
		t.Fatalf("缓存配置被外部修改：%s", config.ListenAddress)
	}
	if starter.getConfig() != config {
		t.Fatal("配置未按 starter 实例缓存")
	}
}
