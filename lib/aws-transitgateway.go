package mpawstgw

import (
  "flag"
  "os"
  "log"
  "time"
  "strings"
  "errors"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/aws/credentials"
  "github.com/aws/aws-sdk-go/aws/credentials/stscreds"
  "github.com/aws/aws-sdk-go/service/cloudwatch"

  mp "github.com/mackerelio/go-mackerel-plugin"
)

// AwsTgwPlugin struct
type AwsTgwPlugin struct {
  Prefix string
  AccessKeyID string
  SecretKeyID string
  Region string
  RoleArn string
  Tgw string
  CloudWatch *cloudwatch.CloudWatch
}

const (
  namespace = "AWS/TransitGateway"
  metricsTypeSum = "Sum"
)

type metrics struct {
  Name string
  Type string
}

// GraphDefinition : return graph definition
func (p AwsTgwPlugin) GraphDefinition() map[string]mp.Graphs {
  labelPrefix := strings.Title(p.Prefix)
  labelPrefix = strings.Replace(labelPrefix, "-", " ", -1)

  // https://docs.aws.amazon.com/vpc/latest/tgw/transit-gateway-cloudwatch-metrics.html#transit-gateway-metrics
  return map[string]mp.Graphs {
    "Bytes": {
      Label: labelPrefix + " Bytes",
      Unit: mp.UnitInteger,
      Metrics: []mp.Metrics{
        {Name: "BytesIn", Label: "Bytes In"},
        {Name: "BytesOut", Label: "Bytes Out"},
      },
    },
      "Packets": {
      Label: labelPrefix + " Packets",
      Unit: mp.UnitInteger,
      Metrics: []mp.Metrics{
        {Name: "PacketsIn", Label: "Packets In"},
        {Name: "PacketsOut", Label: "Packets Out"},
      },
    },
      "Drops": {
      Label: labelPrefix + "Packet Drops",
      Unit: mp.UnitInteger,
      Metrics: []mp.Metrics{
        {Name: "PacketDropCountBlackhole", Label: "Blackhole"},
        {Name: "PacketDropCountNoRoute", Label: "No Route"},
      },
    },
  }
}

// MetricKeyPrefix : interface for PluginWithPrefix
func (p AwsTgwPlugin) MetricKeyPrefix() string {
  if p.Prefix == "" {
    p.Prefix = "TGW"
  }
  return p.Prefix
}

// FetchMetrics : fetch metrics
func (p AwsTgwPlugin) FetchMetrics() (map[string]float64, error) {
  stat := make(map[string]float64)
  for _, met := range []metrics{
    {Name: "BytesIn", Type: metricsTypeSum},
    {Name: "BytesOut", Type: metricsTypeSum},
    {Name: "PacketsIn", Type: metricsTypeSum},
    {Name: "PacketsOut", Type: metricsTypeSum},
    {Name: "PacketDropCountBlackhole", Type: metricsTypeSum},
    {Name: "PacketDropCountNoRoute", Type: metricsTypeSum},
  } {
    v, err := p.getLastPoint(met)
    if err != nil {
      log.Printf("%s : %s", met, err)
    }
    stat[met.Name] = v
  }
  return stat, nil
}

func (p AwsTgwPlugin) getLastPoint(metric metrics) (float64, error) {
  now := time.Now()
  dimensions := []*cloudwatch.Dimension{
    {
      Name: aws.String("TransitGateway"),
      Value: aws.String(p.Tgw),
    },
  }

  response, err := p.CloudWatch.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
    Namespace: aws.String(namespace),
    Dimensions: dimensions,
    StartTime: aws.Time(now.Add(time.Duration(180) * time.Second * -1)), // 3 min (to fetch at least 1 data-point)
    EndTime: aws.Time(now),
    Period: aws.Int64(60),
    MetricName: aws.String(metric.Name),
    Statistics: []*string{aws.String(metric.Type)},
  })
  if err != nil {
    return 0, err
  }

  datapoints := response.Datapoints
  if len(datapoints) == 0 {
    return 0, errors.New("fetch no datapoints : " + p.Tgw)
  }

  // get least recently datapoint.
  // because a most recently datapoint is not stable.
  least := time.Now()
  var latestVal float64
  for _, dp := range datapoints {
    if dp.Timestamp.Before(least) {
      least = *dp.Timestamp
      if metric.Type == metricsTypeSum {
        latestVal = *dp.Sum
      }
    }
  }

  return latestVal, nil
}

func (p *AwsTgwPlugin) prepare() error {
  sess, err := session.NewSession()
  if err != nil {
    return err
  }

  config := aws.NewConfig()
  if p.RoleArn != "" {
    config = config.WithCredentials(stscreds.NewCredentials(sess, p.RoleArn))
  } else if p.AccessKeyID != "" && p.SecretKeyID != "" {
    config = config.WithCredentials(credentials.NewStaticCredentials(p.AccessKeyID, p.SecretKeyID, ""))
  }
  config = config.WithRegion(p.Region)
  p.CloudWatch = cloudwatch.New(sess, config)
  return nil
}

// Do : Do plugin
func Do() {
  optPrefix := flag.String("metric-key-prefix", "", "Metric Key Prefix")
  optAccessKeyID := flag.String("access-key-id", os.Getenv("AWS_ACCESS_KEY_ID"), "AWS Access Key ID")
  optSecretKeyID := flag.String("secret-key-id", os.Getenv("AWS_SECRET_ACCESS_KEY"), "AWS Secret Access Key ID")
  optRegion := flag.String("region", os.Getenv("AWS_DEFAULT_REGION"), "AWS Region")
  optRoleArn := flag.String("role-arn", "", "IAM Role ARN for assume role")
  optTgw := flag.String("tgw", "", "Transit Gateway Resouce ID")
  flag.Parse()

  var AwsTgw AwsTgwPlugin

  AwsTgw.Prefix = *optPrefix
  AwsTgw.AccessKeyID = *optAccessKeyID
  AwsTgw.SecretKeyID = *optSecretKeyID
  AwsTgw.Region = *optRegion
  AwsTgw.RoleArn = *optRoleArn
  AwsTgw.Tgw = *optTgw

  err := AwsTgw.prepare()
  if err != nil {
    log.Fatalln(err)
  }

  helper := mp.NewMackerelPlugin(AwsTgw)
  helper.Run()
}
