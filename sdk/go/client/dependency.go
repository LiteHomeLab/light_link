package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/WQGroup/logger"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

// Dependency 依赖定义
type Dependency struct {
	ServiceName string   // 服务名称
	Methods     []string // 必需的方法列表
}

// DependencyCheckResult 依赖检查结果
type DependencyCheckResult struct {
	ServiceName      string   // 服务名称
	ServiceFound     bool     // 服务是否已注册
	AvailableMethods []string // 可用的方法
	MissingMethods   []string // 缺失的方法
	AllSatisfied     bool     // 是否全部满足
}

// DependencyChecker 依赖检查器
type DependencyChecker struct {
	nc         *nats.Conn
	deps       []Dependency
	registered map[string]*types.ServiceMetadata // 已注册的服务
	mu         sync.RWMutex
	sub        *nats.Subscription
	logger     *logrus.Logger
}

// NewDependencyChecker 创建依赖检查器
func NewDependencyChecker(nc *nats.Conn, deps []Dependency) *DependencyChecker {
	return &DependencyChecker{
		nc:         nc,
		deps:       deps,
		registered: make(map[string]*types.ServiceMetadata),
		logger:     logrus.New(),
	}
}

// GetCheckResults 获取当前检查结果
func (dc *DependencyChecker) GetCheckResults() []*DependencyCheckResult {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	results := make([]*DependencyCheckResult, 0, len(dc.deps))

	for _, dep := range dc.deps {
		result := &DependencyCheckResult{
			ServiceName: dep.ServiceName,
		}

		metadata, exists := dc.registered[dep.ServiceName]
		if exists {
			result.ServiceFound = true

			// 检查每个方法
			available := make([]string, 0)
			missing := make([]string, 0)

			methodMap := make(map[string]bool)
			for _, m := range metadata.Methods {
				methodMap[m.Name] = true
			}

			for _, method := range dep.Methods {
				if methodMap[method] {
					available = append(available, method)
				} else {
					missing = append(missing, method)
				}
			}

			result.AvailableMethods = available
			result.MissingMethods = missing
			result.AllSatisfied = len(missing) == 0
		} else {
			result.ServiceFound = false
			result.MissingMethods = dep.Methods
			result.AllSatisfied = false
		}

		results = append(results, result)
	}

	return results
}

// allSatisfied 检查所有依赖是否满足
func (dc *DependencyChecker) allSatisfied() bool {
	results := dc.GetCheckResults()
	for _, r := range results {
		if !r.AllSatisfied {
			return false
		}
	}
	return true
}

// WaitForDependencies waits for all dependencies to be satisfied
func (dc *DependencyChecker) WaitForDependencies(ctx context.Context) error {
	// Subscribe to registration messages
	sub, err := dc.nc.Subscribe("$LL.register.>", func(msg *nats.Msg) {
		dc.handleRegisterMessage(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	dc.sub = sub
	defer sub.Unsubscribe()

	// Initial progress print
	dc.printInitialRequirements()

	// Wait loop
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			dc.PrintProgress()
		}

		if dc.allSatisfied() {
			dc.printAllSatisfied()
			return nil
		}
	}
}

// handleRegisterMessage handles service registration messages
func (dc *DependencyChecker) handleRegisterMessage(msg *nats.Msg) {
	var registerMsg types.RegisterMessage
	if err := json.Unmarshal(msg.Data, &registerMsg); err != nil {
		dc.logger.Warnf("Failed to unmarshal register message: %v", err)
		return
	}

	if registerMsg.Service == "" {
		dc.logger.Warn("Received register message with empty service name")
		return
	}

	dc.mu.Lock()
	dc.registered[registerMsg.Service] = &registerMsg.Metadata
	dc.mu.Unlock()

	// Print progress update
	dc.PrintProgress()
}

// printInitialRequirements prints initial dependency requirements
func (dc *DependencyChecker) printInitialRequirements() {
	logger.Info("=== Waiting for dependencies ===")
	logger.Infof("Required services: %d", len(dc.deps))

	totalMethods := 0
	for _, dep := range dc.deps {
		logger.Infof("  - %s (%d methods)", dep.ServiceName, len(dep.Methods))
		totalMethods += len(dep.Methods)
	}

	logger.Infof("Total methods required: %d", totalMethods)
	logger.Info("")
}

// PrintProgress prints current dependency check progress
func (dc *DependencyChecker) PrintProgress() {
	results := dc.GetCheckResults()

	totalServices := len(results)
	readyServices := 0
	totalMethods := 0
	availableMethods := 0

	for _, r := range results {
		totalMethods += len(r.AvailableMethods) + len(r.MissingMethods)
		availableMethods += len(r.AvailableMethods)
		if r.AllSatisfied {
			readyServices++
		}
	}

	if totalMethods == 0 {
		return
	}

	// Print overall progress
	logger.Infof("Overall progress: %d/%d methods available (%d/%d services ready)",
		availableMethods, totalMethods, readyServices, totalServices)
	logger.Info("")

	// Print each service status
	for _, r := range results {
		logger.Infof("--- %s ---", r.ServiceName)

		if !r.ServiceFound {
			logger.Infof("  Status: Service not found")
			for _, m := range r.MissingMethods {
				logger.Infof("    ✗ %s (service not registered)", m)
			}
		} else {
			logger.Infof("  Status: %d/%d methods available",
				len(r.AvailableMethods), len(r.AvailableMethods)+len(r.MissingMethods))

			for _, m := range r.AvailableMethods {
				logger.Infof("  ✓ %s", m)
			}
			for _, m := range r.MissingMethods {
				logger.Infof("  ✗ %s (not found)", m)
			}
		}
		logger.Info("")
	}
}

// printAllSatisfied prints all dependencies satisfied message
func (dc *DependencyChecker) printAllSatisfied() {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	logger.Info("=== All dependencies satisfied! ===")
	logger.Info("")
	logger.Info("Available services:")
	logger.Info("")

	results := dc.GetCheckResults()
	for _, r := range results {
		logger.Infof("%s (%d/%d methods)", r.ServiceName,
			len(r.AvailableMethods), len(r.AvailableMethods))

		metadata := dc.registered[r.ServiceName]
		for _, m := range metadata.Methods {
			returns := "void"
			if len(m.Returns) > 0 {
				retList := make([]string, 0, len(m.Returns))
				for _, ret := range m.Returns {
					retList = append(retList, fmt.Sprintf("%s: %s", ret.Name, ret.Type))
				}
				returns = strings.Join(retList, ", ")
			}

			logger.Infof("  -%s %s", m.Name, returns)
		}
		logger.Info("")
	}
}

// Close closes the dependency checker and unsubscribes from NATS.
// This method is safe to call multiple times.
func (dc *DependencyChecker) Close() {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if dc.sub != nil {
		dc.sub.Unsubscribe()
		dc.sub = nil
	}
}
