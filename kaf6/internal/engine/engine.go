package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"

	"kaf6/internal/metrics"
	"kaf6/internal/scenario"
)

type Result struct {
	Name               string
	Description        string
	RunID              string
	Profile            string
	ProfileName        string
	ProfileDescription string
	ProfileSource      string
	ProfileMetricsURL  string
	ConnectivityStatus string
	ConnectivityError  string
	RunError           string
	Brokers            []string
	Produced           int64
	Consumed           int64
	Errors             int64
	ProduceP           metrics.Percentiles
	ConsumeP           metrics.Percentiles
	ConsumePollP       metrics.Percentiles
	Checks             map[string]string
	Duration           time.Duration
	StartedAt          time.Time
	Status             string
}

func Run(ctx context.Context, spec *scenario.ScenarioFile) (*Result, error) {
	start := time.Now()
	runID := start.Format("20060102-150405")
	sum := &metrics.Summary{}
	var runErr error
	runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	connectivityStatus := "ok"
	var connectivityErr error

	verbose := os.Getenv("KAF6_VERBOSE") == "1"
	if verbose {
		fmt.Printf("kaf6 run %s\n", runID)
		fmt.Printf("brokers: %v\n", spec.Brokers)
		fmt.Printf("profile: %s\n", spec.Profile)
	}
	if err := checkConnectivity(spec.Brokers, 2*time.Second); err != nil {
		connectivityStatus = "fail"
		connectivityErr = err
		sum.AddError()
		runErr = fmt.Errorf("connectivity check failed: %w", err)
	}
	if runErr != nil {
		result := &Result{
			Name:               spec.Name,
			Description:        spec.Description,
			RunID:              runID,
			Profile:            spec.Profile,
			ProfileName:        spec.ProfileName,
			ProfileDescription: spec.ProfileDescription,
			ProfileSource:      spec.ProfileSource,
			ProfileMetricsURL:  spec.ProfileMetricsURL,
			ConnectivityStatus: connectivityStatus,
			ConnectivityError:  errorText(connectivityErr),
			RunError:           errorText(runErr),
			Brokers:            spec.Brokers,
			Produced:           sum.Produced,
			Consumed:           sum.Consumed,
			Errors:             sum.Errors,
			ProduceP:           metrics.LatencyPercentiles(sum.ProduceLatencies),
			ConsumeP:           metrics.LatencyPercentiles(sum.ConsumeLatencies),
			ConsumePollP:       metrics.LatencyPercentiles(sum.ConsumePollLatencies),
			Checks:             evaluateChecks(spec, sum),
			Duration:           time.Since(start),
			StartedAt:          start,
			Status:             "fail",
		}
		return result, runErr
	}
	if spec.Scenarios.Producer != nil {
		if err := ensureTopics(runCtx, spec, runID, verbose); err != nil {
			sum.AddError()
			runErr = err
		}
		if runErr != nil {
			result := &Result{
				RunID:              runID,
				Profile:            spec.Profile,
				ProfileName:        spec.ProfileName,
				ProfileDescription: spec.ProfileDescription,
				ProfileSource:      spec.ProfileSource,
				ProfileMetricsURL:  spec.ProfileMetricsURL,
				ConnectivityStatus: connectivityStatus,
				ConnectivityError:  errorText(connectivityErr),
				RunError:           errorText(runErr),
				Brokers:            spec.Brokers,
				Produced:           sum.Produced,
				Consumed:           sum.Consumed,
				Errors:             sum.Errors,
				ProduceP:           metrics.LatencyPercentiles(sum.ProduceLatencies),
				ConsumeP:           metrics.LatencyPercentiles(sum.ConsumeLatencies),
				Checks:             evaluateChecks(spec, sum),
				Duration:           time.Since(start),
				StartedAt:          start,
			}
			if verbose {
				fmt.Printf("result: produced=%d consumed=%d errors=%d\n", result.Produced, result.Consumed, result.Errors)
			}
			return result, runErr
		}
		if verbose {
			topicName := resolveTopic(spec.Scenarios.Producer.Topic, spec.Topics, runID)
			fmt.Printf("scenario: producer (clients=%d messages=%d topic=%s)\n", spec.Scenarios.Producer.Clients, spec.Scenarios.Producer.Messages, topicName)
		}
		if err := runProducer(runCtx, spec, sum, runID); err != nil {
			sum.AddError()
			runErr = err
		}
	}
	if runErr == nil && spec.Scenarios.Consumer != nil {
		time.Sleep(2 * time.Second)
		groupID := resolvedGroupID(spec.Scenarios.Consumer.Group.ID, runID)
		topicName := resolveTopic(spec.Scenarios.Consumer.Topic, spec.Topics, runID)
		if verbose {
			fmt.Printf("scenario: consumer (clients=%d group=%s topic=%s limit=%d)\n", spec.Scenarios.Consumer.Clients, groupID, topicName, spec.Scenarios.Consumer.Limit)
		}
		if err := runConsumer(runCtx, spec, sum, runID); err != nil {
			sum.AddError()
			runErr = err
		}
	}
	if runErr == nil && spec.Scenarios.Metrics != nil {
		if verbose {
			fmt.Printf("scenario: metrics (url=%s)\n", spec.Scenarios.Metrics.URL)
		}
		if err := runMetrics(runCtx, spec, sum); err != nil {
			sum.AddError()
			runErr = err
		}
	}

	checks := evaluateChecks(spec, sum)

	result := &Result{
		Name:               spec.Name,
		Description:        spec.Description,
		RunID:              runID,
		Profile:            spec.Profile,
		ProfileName:        spec.ProfileName,
		ProfileDescription: spec.ProfileDescription,
		ProfileSource:      spec.ProfileSource,
		ProfileMetricsURL:  spec.ProfileMetricsURL,
		ConnectivityStatus: connectivityStatus,
		ConnectivityError:  errorText(connectivityErr),
		RunError:           errorText(runErr),
		Brokers:            spec.Brokers,
		Produced:           sum.Produced,
		Consumed:           sum.Consumed,
		Errors:             sum.Errors,
		ProduceP:           metrics.LatencyPercentiles(sum.ProduceLatencies),
		ConsumeP:           metrics.LatencyPercentiles(sum.ConsumeLatencies),
		ConsumePollP:       metrics.LatencyPercentiles(sum.ConsumePollLatencies),
		Checks:             checks,
		Duration:           time.Since(start),
		StartedAt:          start,
	}
	result.Status = "pass"
	if runErr != nil {
		result.Status = "fail"
	} else {
		for _, status := range checks {
			if status != "pass" {
				result.Status = "fail"
				break
			}
		}
	}
	if result.Status == "pass" && sum.Errors > 0 {
		result.Status = "fail"
	}
	if result.Status == "pass" && connectivityStatus != "ok" {
		result.Status = "fail"
	}
	if verbose {
		fmt.Printf("result: produced=%d consumed=%d errors=%d\n", result.Produced, result.Consumed, result.Errors)
	}
	return result, runErr
}

func runProducer(ctx context.Context, spec *scenario.ScenarioFile, sum *metrics.Summary, runID string) error {
	cfg := spec.Scenarios.Producer
	if cfg.Clients <= 0 {
		cfg.Clients = 1
	}
	if cfg.Messages <= 0 {
		cfg.Messages = 1
	}
	topic := resolveTopic(cfg.Topic, spec.Topics, runID)
	if topic == "" {
		return fmt.Errorf("producer topic is required")
	}

	options := []kgo.Opt{
		kgo.SeedBrokers(spec.Brokers...),
		kgo.DisableIdempotentWrite(),
		kgo.AllowAutoTopicCreation(),
	}
	if os.Getenv("KAF6_DEBUG") != "0" {
		options = append(options, kgo.WithLogger(newDebugLogger("producer")))
	}
	client, err := kgo.NewClient(options...)
	if err != nil {
		return err
	}
	defer client.Close()

	payloadTemplate := cfg.Value.JSON
	total := cfg.Messages
	perClient := total / cfg.Clients
	if perClient == 0 {
		perClient = 1
	}

	var wg sync.WaitGroup
	wg.Add(cfg.Clients)
	for i := 0; i < cfg.Clients; i++ {
		go func(clientID int) {
			defer wg.Done()
			for j := 0; j < perClient; j++ {
				select {
				case <-ctx.Done():
					sum.AddError()
					return
				default:
				}
				value, err := buildPayload(payloadTemplate)
				if err != nil {
					sum.AddError()
					continue
				}
				start := time.Now()
				res := client.ProduceSync(ctx, &kgo.Record{
					Topic: topic,
					Value: value,
				})
				if err := res.FirstErr(); err != nil {
					sum.AddError()
					if os.Getenv("KAF6_VERBOSE") == "1" {
						fmt.Printf("producer[%d]: error: %v\n", clientID, err)
					}
					continue
				}
				sum.AddProduce(time.Since(start))
				if os.Getenv("KAF6_VERBOSE") == "1" && (sum.Produced%10) == 0 {
					fmt.Printf("producer: sent=%d\n", sum.Produced)
				}
			}
		}(i)
	}
	wg.Wait()
	return nil
}

func runConsumer(ctx context.Context, spec *scenario.ScenarioFile, sum *metrics.Summary, runID string) error {
	cfg := spec.Scenarios.Consumer
	if cfg.Clients <= 0 {
		cfg.Clients = 1
	}
	if cfg.Limit <= 0 {
		cfg.Limit = 1
	}
	groupID := resolvedGroupID(cfg.Group.ID, runID)
	topic := resolveTopic(cfg.Topic, spec.Topics, runID)
	if topic == "" {
		return fmt.Errorf("consumer topic is required")
	}
	timeout := 30 * time.Second
	if cfg.Timeout != "" {
		if parsed, err := time.ParseDuration(cfg.Timeout); err == nil {
			timeout = parsed
		}
	}

	options := []kgo.Opt{
		kgo.SeedBrokers(spec.Brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic),
		kgo.DisableIdempotentWrite(),
		kgo.BlockRebalanceOnPoll(),
		kgo.AllowAutoTopicCreation(),
	}
	if cfg.Offset == "" || cfg.Offset == "earliest" {
		options = append(options, kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()))
	}
	if os.Getenv("KAF6_DEBUG") != "0" {
		options = append(options, kgo.WithLogger(newDebugLogger("consumer")))
	}
	client, err := kgo.NewClient(options...)
	if err != nil {
		return err
	}

	limit := cfg.Limit
	consumed := 0
	deadline := time.Now().Add(timeout)
	debug := os.Getenv("KAF6_DEBUG") == "1"
	idleDeadline := time.Now().Add(15 * time.Second)
	if debug {
		fmt.Printf("consumer debug: group=%s topic=%s limit=%d timeout=%s\n", groupID, topic, limit, timeout)
	}

	for consumed < limit {
		if time.Now().After(deadline) {
			if debug {
				fmt.Printf("consumer debug: deadline exceeded after %s\n", timeout)
			}
			break
		}
		if time.Now().After(idleDeadline) {
			if debug {
				fmt.Printf("consumer debug: idle timeout reached\n")
			}
			break
		}
		pollStart := time.Now()
		pollCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		fetches := client.PollFetches(pollCtx)
		cancel()
		pollLatency := time.Since(pollStart)
		if pollCtx.Err() == context.DeadlineExceeded {
			continue
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			sum.AddError()
			if debug {
				fmt.Printf("consumer debug: fetch errors: %+v\n", errs)
			}
			return fmt.Errorf("fetch errors: %+v", errs)
		}
		fetches.EachRecord(func(record *kgo.Record) {
			consumed++
			sum.AddConsumePoll(pollLatency)
			sum.AddConsume(time.Since(record.Timestamp))
			if os.Getenv("KAF6_VERBOSE") == "1" && (sum.Consumed%10) == 0 {
				fmt.Printf("consumer: received=%d\n", sum.Consumed)
			}
			if debug {
				fmt.Printf("consumer debug: record topic=%s partition=%d offset=%d\n", record.Topic, record.Partition, record.Offset)
			}
			idleDeadline = time.Now().Add(15 * time.Second)
		})
	}
	if consumed < limit {
		closeClient(client, debug)
		return fmt.Errorf("consume timeout: got %d of %d", consumed, limit)
	}
	closeClient(client, debug)
	return nil
}

func buildPayload(template map[string]string) ([]byte, error) {
	if len(template) == 0 {
		return []byte(fmt.Sprintf(`{"uuid":"%d"}`, time.Now().UnixNano())), nil
	}
	payload := make(map[string]string, len(template))
	for key, val := range template {
		switch val {
		case "{{uuid}}":
			payload[key] = fmt.Sprintf("%d", time.Now().UnixNano())
		case "{{now}}":
			payload[key] = time.Now().UTC().Format(time.RFC3339Nano)
		default:
			payload[key] = val
		}
	}
	return json.Marshal(payload)
}

func replaceRunID(input string, runID string) string {
	return strings.ReplaceAll(input, "{{run_id}}", runID)
}

func resolvedGroupID(input string, runID string) string {
	if input == "" || input == "{{run_id}}" {
		return "kaf6-group-" + runID
	}
	return replaceRunID(input, runID)
}

func resolveTopic(input string, topics []scenario.TopicSpec, runID string) string {
	name := input
	if name == "" && len(topics) > 0 {
		name = topics[0].Name
	}
	if runID == "" {
		return name
	}
	return replaceRunID(name, runID)
}

func ensureTopics(ctx context.Context, spec *scenario.ScenarioFile, runID string, verbose bool) error {
	if len(spec.Topics) == 0 {
		return nil
	}
	client, err := kgo.NewClient(kgo.SeedBrokers(spec.Brokers...))
	if err != nil {
		return err
	}
	defer client.Close()
	admin := kadm.NewClient(client)

	for _, topic := range spec.Topics {
		name := resolveTopic(topic.Name, nil, runID)
		if name == "" {
			continue
		}
		if topic.Recreate {
			if verbose {
				fmt.Printf("topic: deleting %s\n", name)
			}
			_, _ = admin.DeleteTopics(ctx, name)
		}
		if verbose {
			fmt.Printf("topic: creating %s (partitions=%d)\n", name, topic.Partitions)
		}
		partitions := topic.Partitions
		if partitions <= 0 {
			partitions = 1
		}
		_, err := admin.CreateTopics(ctx, partitions, 1, nil, name)
		if err != nil {
			return fmt.Errorf("create topic %s: %w", name, err)
		}
	}
	time.Sleep(2 * time.Second)
	return nil
}

func closeClient(client *kgo.Client, debug bool) {
	done := make(chan struct{})
	go func() {
		client.CloseAllowingRebalance()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		if debug {
			fmt.Printf("consumer debug: client close timed out\n")
		}
	}
}

func runMetrics(ctx context.Context, spec *scenario.ScenarioFile, sum *metrics.Summary) error {
	cfg := spec.Scenarios.Metrics
	if cfg.URL == "" {
		return fmt.Errorf("metrics url is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("metrics status %d", resp.StatusCode)
	}
	return nil
}

type debugLogger struct {
	prefix string
}

func newDebugLogger(prefix string) *debugLogger {
	return &debugLogger{prefix: prefix}
}

func (l *debugLogger) Log(level kgo.LogLevel, msg string, keyvals ...interface{}) {
	parts := make([]string, 0, len(keyvals)/2)
	for i := 0; i+1 < len(keyvals); i += 2 {
		parts = append(parts, fmt.Sprintf("%v=%v", keyvals[i], keyvals[i+1]))
	}
	if len(parts) > 0 {
		fmt.Printf("%s [%s] %s %s\n", l.prefix, level, msg, strings.Join(parts, " "))
		return
	}
	fmt.Printf("%s [%s] %s\n", l.prefix, level, msg)
}

func (l *debugLogger) Level() kgo.LogLevel {
	return kgo.LogLevelDebug
}

func evaluateChecks(spec *scenario.ScenarioFile, sum *metrics.Summary) map[string]string {
	out := make(map[string]string)
	for _, check := range spec.Checks {
		switch check.Type {
		case "count_equals":
			actual := int(sum.Consumed)
			if check.Metric == "produced" {
				actual = int(sum.Produced)
			}
			if actual == check.Expected {
				out[check.Name] = "pass"
			} else {
				out[check.Name] = "fail"
			}
		default:
			out[check.Name] = "skip"
		}
	}
	return out
}

func checkConnectivity(brokers []string, timeout time.Duration) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no brokers configured")
	}
	var lastErr error
	for _, broker := range brokers {
		conn, err := net.DialTimeout("tcp", broker, timeout)
		if err != nil {
			lastErr = err
			continue
		}
		_ = conn.Close()
		return nil
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("unable to connect to any broker")
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
