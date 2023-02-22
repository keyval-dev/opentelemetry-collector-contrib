// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package odigosresourcenameprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/odigosresourcenameprocessor"

import (
	"context"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.uber.org/zap"
	"sync"
)

const (
	// The value of "type" key in configuration.
	typeStr = "odigosresourcename"
	// The stability level of the processor.
	stability = component.StabilityLevelBeta
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}
var nameResolver *NameResolver

// NewFactory returns a new factory for the Resource processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, stability),
		processor.WithMetrics(createMetricsProcessor, stability),
		processor.WithLogs(createLogsProcessor, stability))
}

func createDefaultConfig() component.Config {
	return &Config{
		APIConfig: k8sconfig.APIConfig{AuthType: k8sconfig.AuthTypeServiceAccount},
	}
}

func createTracesProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Traces) (processor.Traces, error) {

	if err := initNameResolver(cfg, set.Logger); err != nil {
		return nil, err
	}

	proc := &resourceProcessor{logger: set.Logger, nameResolver: nameResolver}
	return processorhelper.NewTracesProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processTraces,
		processorhelper.WithCapabilities(processorCapabilities))
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics) (processor.Metrics, error) {

	if err := initNameResolver(cfg, set.Logger); err != nil {
		return nil, err
	}

	proc := &resourceProcessor{logger: set.Logger, nameResolver: nameResolver}
	return processorhelper.NewMetricsProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processMetrics,
		processorhelper.WithCapabilities(processorCapabilities))
}

func createLogsProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Logs) (processor.Logs, error) {

	if err := initNameResolver(cfg, set.Logger); err != nil {
		return nil, err
	}

	proc := &resourceProcessor{logger: set.Logger, nameResolver: nameResolver}
	return processorhelper.NewLogsProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processLogs,
		processorhelper.WithCapabilities(processorCapabilities))
}

func initNameResolver(cfg component.Config, logger *zap.Logger) error {
	if nameResolver != nil {
		return nil
	}

	pCfg := cfg.(*Config)
	kubeClient, err := k8sconfig.MakeClient(pCfg.APIConfig)
	if err != nil {
		return err
	}

	kubelet, err := NewKubeletClient()
	if err != nil {
		return err
	}

	nameResolver = &NameResolver{
		kc:            kubeClient,
		logger:        logger,
		devicesToPods: map[string]string{},
		mu:            sync.RWMutex{},
		kubelet:       kubelet,
	}
	return nameResolver.Start()
}
