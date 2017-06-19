# KafkaLogrus

kafka hook for logrus

## Usage

```go
addrs := []string{"localhost:9092"}

log := logrus.New()

hook, err := kafkahook.NewHook(addrs, "log", nil)
if err == nil {
  log.Hooks.Add(hook)
}
log.WithFields(logrus.Fields{
  "foo":  "bar",
  "number": 42,
}).Infoln("found the answer")
```
