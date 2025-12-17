# Glyph Deployment Guide

This guide covers deploying Glyph applications to production environments using Docker, Kubernetes, and cloud providers.

## Table of Contents

1. [Docker Deployment](#docker-deployment)
2. [Kubernetes Deployment](#kubernetes-deployment)
3. [Cloud Provider Deployment](#cloud-provider-deployment)
4. [Monitoring & Observability](#monitoring--observability)
5. [Best Practices](#best-practices)

---

## Docker Deployment

### Quick Start with Docker Compose

```bash
# Start all services (Glyph app + PostgreSQL + Redis + Monitoring)
docker-compose up -d

# View logs
docker-compose logs -f glyph

# Stop services
docker-compose down
```

### Production Docker Build

```bash
# Build production image
docker build -t glyph:latest .

# Run container
docker run -d \
  --name glyph-app \
  -p 8080:8080 \
  -e Glyph_ENV=production \
  -e DATABASE_URL=postgresql://user:pass@postgres:5432/db \
  glyph:latest
```

### Development with Hot-Reload

```bash
# Build development image
docker build -f Dockerfile.dev -t glyph:dev .

# Run with hot-reload
docker run -it \
  --name glyph-dev \
  -p 8080:8080 \
  -v $(pwd):/app \
  glyph:dev
```

### Docker Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `Glyph_ENV` | Environment (development/production) | production |
| `Glyph_PORT` | HTTP server port | 8080 |
| `Glyph_LOG_LEVEL` | Log level (debug/info/warn/error) | info |
| `DATABASE_URL` | PostgreSQL connection string | - |
| `REDIS_URL` | Redis connection string | - |

---

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured
- Helm 3 (optional)

### Deploy to Kubernetes

```bash
# Create namespace
kubectl apply -f deploy/kubernetes/namespace.yaml

# Deploy PostgreSQL
kubectl apply -f deploy/kubernetes/postgres.yaml

# Deploy Redis
kubectl apply -f deploy/kubernetes/redis.yaml

# Deploy Glyph application
kubectl apply -f deploy/kubernetes/deployment.yaml

# Create ingress (optional)
kubectl apply -f deploy/kubernetes/ingress.yaml
```

### Verify Deployment

```bash
# Check pods
kubectl get pods -n glyph

# Check services
kubectl get svc -n glyph

# View logs
kubectl logs -f deployment/glyph-app -n glyph

# Check HPA (autoscaling)
kubectl get hpa -n glyph
```

### Scaling

```bash
# Manual scaling
kubectl scale deployment glyph-app --replicas=5 -n glyph

# Autoscaling is configured via HPA (70% CPU, 3-10 replicas)
kubectl get hpa glyph-hpa -n glyph
```

### Rolling Updates

```bash
# Update image
kubectl set image deployment/glyph-app glyph=glyph:v2.0 -n glyph

# Check rollout status
kubectl rollout status deployment/glyph-app -n glyph

# Rollback if needed
kubectl rollout undo deployment/glyph-app -n glyph
```

---

## Cloud Provider Deployment

### AWS (EKS)

```bash
# Create EKS cluster
eksctl create cluster \
  --name glyph-cluster \
  --region us-east-1 \
  --nodegroup-name glyph-nodes \
  --node-type t3.medium \
  --nodes 3 \
  --nodes-min 3 \
  --nodes-max 10

# Configure kubectl
aws eks update-kubeconfig --name glyph-cluster --region us-east-1

# Deploy application
kubectl apply -f deploy/kubernetes/
```

### Google Cloud (GKE)

```bash
# Create GKE cluster
gcloud container clusters create glyph-cluster \
  --num-nodes=3 \
  --machine-type=n1-standard-2 \
  --enable-autoscaling \
  --min-nodes=3 \
  --max-nodes=10 \
  --region=us-central1

# Get credentials
gcloud container clusters get-credentials glyph-cluster --region=us-central1

# Deploy application
kubectl apply -f deploy/kubernetes/
```

### Azure (AKS)

```bash
# Create AKS cluster
az aks create \
  --resource-group glyph-rg \
  --name glyph-cluster \
  --node-count 3 \
  --enable-addons monitoring \
  --generate-ssh-keys

# Get credentials
az aks get-credentials --resource-group glyph-rg --name glyph-cluster

# Deploy application
kubectl apply -f deploy/kubernetes/
```

---

## Monitoring & Observability

### Prometheus Metrics

Glyph exposes metrics at `/metrics`:

```
# Request metrics
http_requests_total
http_request_duration_seconds
http_requests_in_flight

# System metrics
process_cpu_seconds_total
process_resident_memory_bytes
go_goroutines

# Application metrics
glyph_compilation_duration_seconds
glyph_route_executions_total
glyph_database_queries_total
```

### Grafana Dashboards

```bash
# Access Grafana (port-forward)
kubectl port-forward svc/grafana 3000:3000 -n glyph

# Login: admin / admin
# Pre-configured dashboard at: Glyph Application Metrics
```

### Log Aggregation

```bash
# View application logs
kubectl logs -f deployment/glyph-app -n glyph

# Stream logs from all pods
kubectl logs -f -l app=glyph -n glyph

# Export logs to file
kubectl logs deployment/glyph-app -n glyph > glyph-logs.txt
```

### Health Checks

```bash
# Liveness probe
curl http://localhost:8080/health

# Readiness probe
curl http://localhost:8080/ready

# Metrics endpoint
curl http://localhost:8080/metrics
```

---

## Best Practices

### Security

1. **Use non-root containers** ✅ (UID 1000)
2. **Store secrets in Kubernetes Secrets** ✅
3. **Enable TLS/HTTPS** (configure ingress)
4. **Regular security scanning**
   ```bash
   docker scan glyph:latest
   ```

### Performance

1. **Resource Limits**
   - CPU: 100m-500m
   - Memory: 128Mi-512Mi
   - Adjust based on load testing

2. **Connection Pooling**
   - Database: Max 20 connections
   - Redis: Max 10 connections

3. **Caching**
   - Use Redis for session/cache
   - Enable HTTP caching headers

### High Availability

1. **Multi-replica deployment** (min 3 replicas)
2. **Pod Anti-Affinity** (spread across nodes)
3. **Health checks** (liveness + readiness)
4. **Graceful shutdown** (30s timeout)
5. **Horizontal Pod Autoscaling** (70% CPU threshold)

### Disaster Recovery

1. **Database Backups**
   ```bash
   # Backup PostgreSQL
   kubectl exec -n glyph glyph-postgres-0 -- \
     pg_dump -U glyph_user glyph_db > backup.sql
   ```

2. **Persistent Volume Snapshots**
   ```bash
   kubectl get pvc -n glyph
   # Create snapshots via cloud provider
   ```

3. **Configuration Backup**
   ```bash
   kubectl get all,cm,secret -n glyph -o yaml > glyph-backup.yaml
   ```

### Monitoring Checklist

- [ ] Prometheus scraping all endpoints
- [ ] Grafana dashboards configured
- [ ] Alerting rules defined
- [ ] Log aggregation enabled
- [ ] Error tracking configured
- [ ] Performance baselines established

---

## Troubleshooting

### Pod not starting

```bash
# Check pod events
kubectl describe pod <pod-name> -n glyph

# Check logs
kubectl logs <pod-name> -n glyph

# Check resource limits
kubectl top pod <pod-name> -n glyph
```

### Database connection issues

```bash
# Test database connectivity
kubectl exec -it deployment/glyph-app -n glyph -- \
  psql -h glyph-postgres -U glyph_user -d glyph_db

# Check database logs
kubectl logs statefulset/glyph-postgres -n glyph
```

### High memory usage

```bash
# Check memory metrics
kubectl top pods -n glyph

# Increase memory limits
kubectl set resources deployment glyph-app \
  --limits=memory=1Gi -n glyph
```

---

## Production Checklist

Before going to production:

- [ ] Load testing completed (>1000 req/s)
- [ ] Security scan passed
- [ ] Database migrations tested
- [ ] Backups configured and tested
- [ ] Monitoring and alerting active
- [ ] TLS/HTTPS configured
- [ ] Rate limiting enabled
- [ ] Error tracking configured
- [ ] Runbooks documented
- [ ] On-call rotation established

---

## Support

For deployment issues:
- Check logs: `kubectl logs -f deployment/glyph-app -n glyph`
- Check metrics: Access Grafana dashboard
- Review documentation: `docs/`
- Open issue: GitHub issues

---

*Last Updated: 2025-12-09*
*Glyph Version: 1.0.0-production*
