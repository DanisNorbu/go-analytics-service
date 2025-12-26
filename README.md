# Go Analytics Service

Простой high-load сервис на Go для приёма метрик, расчёта скользящего среднего и детекции аномалий по z-score, с кэшированием в Redis и мониторингом через Prometheus/Grafana. Развёрнут в Kubernetes (Minikube) с автоскейлингом по CPU (HPA).

## Основной функционал

* HTTP API для потоковых метрик:

  * `POST /ingest` — приём метрики

    ```json
    {
      "timestamp": 1735100000,
      "cpu": 60.5,
      "rps": 2000
    }
    ```
  * `GET /stats` — агрегированная статистика:

    * количество метрик,
    * средний CPU / RPS (rolling average по окну),
    * последний RPS / timestamp,
    * z-score и флаг аномалии.
  * `GET /health` — простая проверка живости.
  * `GET /metrics` — метрики Prometheus.

* Статистическая аналитика:

  * скользящее среднее (rolling average) по окну последних N событий;
  * детекция аномалий по z-score (порог > 2σ);
  * обработка в отдельной горутине через канал метрик.

* Инфраструктура:

  * Redis для хранения последних метрик;
  * Kubernetes: Deployment + Service + HPA (2–5 реплик по CPU);
  * Prometheus + Grafana для мониторинга (RPS, latency, anomaly rate).

## Стек

* Go `1.25.x`
* Redis `7`
* Docker / Docker Desktop
* Minikube
* Kubernetes (Deployment, Service, HPA)
* Prometheus, Grafana (через Helm)
* Нагрузочное тестирование: `hey`

## Запуск локально (без Kubernetes)

Требуется запущенный Redis в Docker:

```powershell
docker run -d --name redis -p 6379:6379 redis:7
```

Сервис:

```powershell
go mod tidy
go run .
```

Проверка:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
```

Пример отправки метрики:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/ingest" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"timestamp":0,"cpu":50,"rps":1500}'
```

## Docker

Сборка образа:

```powershell
docker build -t go-analytics-service:latest .
```

(Размер образа ~30–40 MB, что укладывается в требование `< 300MB`.)

## Kubernetes (Minikube)

1. Запуск Minikube:

```powershell
minikube start --driver=docker --cpus=2 --memory=4096
```

2. Переключение Docker на Minikube и сборка образа внутри кластера:

```powershell
& minikube -p minikube docker-env --shell powershell | Invoke-Expression
docker build -t go-analytics-service:latest .
```

3. Применение манифестов:

```powershell
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/go-analytics-service.yaml
kubectl apply -f k8s/hpa.yaml
```

4. Доступ к сервису из локальной машины:

```powershell
kubectl port-forward svc/go-analytics-service 8080:8080
```

Теперь API доступно по `http://localhost:8080`.

## Prometheus и Grafana

Установка в namespace `monitoring`:

```powershell
kubectl create namespace monitoring

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

helm install prometheus prometheus-community/prometheus --namespace monitoring
helm install grafana grafana/grafana --namespace monitoring
```

Доступ:

```powershell
# Prometheus
kubectl port-forward -n monitoring svc/prometheus-server 9090:80

# Grafana
kubectl port-forward -n monitoring svc/grafana 3000:80
```

## Нагрузочное тестирование

Пример с `hey` (около 2000+ RPS):

```powershell
hey -z 10s -c 20 `
  -m POST `
  -H "Content-Type: application/json" `
  -d '{"timestamp":0,"cpu":60,"rps":2000}' `
  http://localhost:8080/ingest
```

Для демонстрации HPA можно запустить более долгий тест (например, 300 секунд и 100 соединений) и параллельно смотреть:

```powershell
kubectl get hpa -w
kubectl get pods -w
```

---
