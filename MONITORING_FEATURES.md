# 🚀 Sistema de Monitoramento e Alertas - Aegis Backup v0.6

Este documento descreve as novas funcionalidades de monitoramento e alertas implementadas no Aegis Backup.

## 🎯 **Novas Funcionalidades**

### 1. **Sistema de Métricas Avançado**
- **Coleta automática** de métricas de backup para cada dispositivo
- **Status em tempo real** (Healthy, Warning, Critical, Unknown)
- **Estatísticas detalhadas**:
  - Tempo médio de backup
  - Taxa de sucesso/falha
  - Tamanho dos backups
  - Falhas consecutivas
  - Último backup realizado

### 2. **Sistema de Alertas Inteligentes**
- **Alertas automáticos** baseados em regras configuráveis
- **Regras padrão incluídas**:
  - 3+ falhas consecutivas (Critical)
  - Backup demorando mais de 5 minutos (Warning)
  - Sem backup há mais de 24h (Critical)
  - Mudança significativa no tamanho do backup (Warning)
- **Cooldown configurável** para evitar spam de alertas
- **Acknowledgment** e limpeza de alertas

### 3. **API REST Completa**
- **Endpoints disponíveis**:
  - `GET /health` - Health check
  - `GET /api/v1/metrics` - Todas as métricas
  - `GET /api/v1/metrics/{device}` - Métricas específicas
  - `GET /api/v1/stats` - Estatísticas gerais
  - `GET /api/v1/alerts` - Alertas ativos
  - `POST /api/v1/alerts/{id}` - Reconhecer alerta
  - `DELETE /api/v1/alerts/{id}` - Limpar alerta
  - `GET /` - Dashboard web

### 4. **Dashboard Web Integrado**
- **Interface visual** para monitoramento
- **Estatísticas em tempo real**:
  - Total de dispositivos
  - Dispositivos saudáveis/em aviso/críticos
  - Total de backups e falhas
- **Lista de alertas ativos** com severidade
- **Auto-refresh** a cada 30 segundos
- **Design responsivo** e moderno

## 🛠️ **Como Usar**

### **Iniciando com Monitoramento**
```bash
# Modo único com API na porta padrão (8080)
./main -config config.json -devices devices.csv

# Modo daemon com API personalizada
./main -daemon -api-port 9090

# Verificar se está funcionando
curl http://localhost:8080/health
```

### **Acessando o Dashboard**
1. Abra o navegador em `http://localhost:8080`
2. Visualize o status de todos os dispositivos
3. Monitore alertas em tempo real
4. Acompanhe estatísticas gerais

### **Usando a API**
```bash
# Obter métricas de todos os dispositivos
curl http://localhost:8080/api/v1/metrics

# Obter métricas de um dispositivo específico
curl http://localhost:8080/api/v1/metrics/router1

# Obter estatísticas gerais
curl http://localhost:8080/api/v1/stats

# Listar alertas ativos
curl http://localhost:8080/api/v1/alerts

# Reconhecer um alerta
curl -X POST http://localhost:8080/api/v1/alerts/alert_id

# Limpar um alerta
curl -X DELETE http://localhost:8080/api/v1/alerts/alert_id
```

## 📊 **Estrutura de Dados**

### **Métricas do Dispositivo**
```json
{
  "device_name": "router1",
  "last_backup": "2024-01-15T10:30:00Z",
  "last_success": "2024-01-15T10:30:00Z",
  "last_failure": "2024-01-14T15:20:00Z",
  "backup_size": 1024,
  "success_count": 45,
  "failure_count": 2,
  "average_time": "2m30s",
  "last_error": "connection timeout",
  "status": "healthy",
  "consecutive_failures": 0
}
```

### **Estatísticas Gerais**
```json
{
  "total_devices": 5,
  "healthy_devices": 4,
  "warning_devices": 1,
  "critical_devices": 0,
  "unknown_devices": 0,
  "total_backups": 225,
  "total_failures": 8,
  "average_backup_time": "2m15s"
}
```

### **Alerta Ativo**
```json
{
  "id": "router1_consecutive_failures_1705312200",
  "device_name": "router1",
  "rule_name": "consecutive_failures",
  "severity": "critical",
  "message": "Device router1 has failed 3 consecutive backups. Last error: connection timeout",
  "timestamp": "2024-01-15T10:30:00Z",
  "channels": ["telegram"],
  "acknowledged": false
}
```

## 🔧 **Configuração**

### **Porta da API**
```bash
# Alterar porta padrão (8080)
./main -api-port 9090
```

### **Regras de Alerta Personalizadas**
As regras de alerta são configuráveis via código. Exemplos de regras disponíveis:

- `consecutive_failures >= 3` - Falhas consecutivas
- `backup_time > 5m` - Backup demorado
- `last_backup_age > 24h` - Sem backup há muito tempo
- `status == critical` - Status crítico

## 🚨 **Tipos de Status**

- **🟢 Healthy**: Dispositivo funcionando normalmente
- **🟡 Warning**: Algum problema detectado (1-2 falhas consecutivas)
- **🔴 Critical**: Problema sério (3+ falhas consecutivas)
- **⚪ Unknown**: Status inicial ou sem dados suficientes

## 📈 **Benefícios**

1. **Visibilidade Total**: Monitore todos os dispositivos em uma única interface
2. **Alertas Proativos**: Seja notificado imediatamente sobre problemas
3. **Métricas Históricas**: Acompanhe tendências e performance
4. **API Integrável**: Integre com outros sistemas de monitoramento
5. **Dashboard Visual**: Interface amigável para operadores
6. **Escalabilidade**: Suporta centenas de dispositivos

## 🔄 **Integração com Sistemas Existentes**

### **Prometheus/Grafana**
```yaml
# Exemplo de configuração Prometheus
scrape_configs:
  - job_name: 'aegis-backup'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/api/v1/metrics'
```

### **Webhooks**
```bash
# Exemplo de webhook para alertas
curl -X POST http://your-webhook-url \
  -H "Content-Type: application/json" \
  -d '{"alert": "critical", "device": "router1"}'
```

## 🎯 **Próximos Passos Sugeridos**

1. **Notificações por Email**: Integração com SMTP
2. **Slack Integration**: Alertas diretos no Slack
3. **Métricas Prometheus**: Exportação nativa de métricas
4. **Dashboard Avançado**: Gráficos e históricos
5. **Backup Diferencial**: Detecção de mudanças
6. **Health Checks**: Verificação proativa de conectividade

## 📝 **Arquivos Modificados/Criados**

### **Novos Módulos**:
- `internal/monitoring/metrics.go` - Sistema de métricas
- `internal/monitoring/alerts.go` - Sistema de alertas
- `internal/monitoring/metrics_test.go` - Testes
- `internal/api/server.go` - API REST e dashboard

### **Modificados**:
- `cmd/main.go` - Integração do monitoramento
- `internal/backup/backup.go` - Coleta de métricas
- `internal/worker/pool.go` - Passagem de parâmetros
- `internal/scheduler/scheduler.go` - Integração com scheduler

### **Documentação**:
- `MONITORING_FEATURES.md` - Este arquivo

---

**🎉 O Aegis Backup agora é uma solução completa de backup com monitoramento profissional!**
