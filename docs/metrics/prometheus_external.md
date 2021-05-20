```yaml
apiVersion: autoscaling/v2beta2
kind: HorizontalPodAutoscaler
metadata:
  name: prometheus-hpa
  annotations:
    "prometheus.server": "http://cn-beijing.arms.aliyuncs.com:9090/api/v1/prometheus/729f125f18d91f4a17f6607d6eb191/1845311666427154/cc689df2d13e24f40a87c775b9cd8a0bc/cn-beijing"
    "prometheus.query": "up"
    "prometheus.metric.name": "carson_metric"
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: nginx-deployment-basic
  minReplicas: 1
  maxReplicas: 10
  metrics:
    - type: External
      external:
        metric:
          name: carson_metric
        target:
          type: Value
          value: 5
```
      