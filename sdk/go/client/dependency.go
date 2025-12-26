package client

import (
	"sync"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
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
