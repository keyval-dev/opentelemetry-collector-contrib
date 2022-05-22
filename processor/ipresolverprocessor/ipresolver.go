package ipresolverprocessor

import (
	"context"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/otel/semconv/v1.6.1"
	"go.uber.org/zap"
	"net"
)

type ipResolverProcessor struct {
	logger   *zap.Logger
}

func newIPResolverProcessor(logger *zap.Logger) *ipResolverProcessor {
	return &ipResolverProcessor{
		logger: logger,
	}
}

func (ir *ipResolverProcessor) processTraces(_ context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	for i := 0; i < td.ResourceSpans().Len(); i++ {
		ss := td.ResourceSpans().At(i).ScopeSpans()
		for j := 0; j < ss.Len(); j++ {
			spans := ss.At(j).Spans()
			for x := 0; x < spans.Len(); x++ {
				span := spans.At(x)
				netPeerName, netPeerNameExists := span.Attributes().Get(string(semconv.NetPeerNameKey))
				// If no net.peer.name or net.peer.name is an ip
				if !netPeerNameExists || net.ParseIP(netPeerName.StringVal()) != nil {
					var ip net.IP
					if netPeerIP, exists := span.Attributes().Get(string(semconv.NetPeerIPKey)); exists {
						ip = net.ParseIP(netPeerIP.StringVal())
					} else if netPeerNameExists {
						ip = net.ParseIP(netPeerName.StringVal())
					}

					if ip != nil {
						hosts, err := net.LookupAddr(ip.String())
						if err != nil {
							ir.logger.Error("could not find hostname for net.host.ip", zap.Error(err))
						} else if len(hosts) > 0 {
							if netPeerNameExists {
								span.Attributes().UpdateString(string(semconv.NetPeerNameKey), hosts[0])
							} else {
								span.Attributes().InsertString(string(semconv.NetPeerNameKey), hosts[0])
							}
						}
					}
				}
			}
		}
	}

	return td, nil
}
