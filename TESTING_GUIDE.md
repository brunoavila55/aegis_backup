# 🧪 Guia de Testes - Aegis Backup v0.6

Este guia mostra como testar todas as funcionalidades do sistema de monitoramento e alertas.

## 🚀 **Teste Rápido (5 minutos)**

### 1. **Compilar e Executar**
```bash
# Compilar o projeto
go build ./cmd/main.go

# Executar com monitoramento
./main -config config.json -devices devices.csv
```

### 2. **Verificar se está funcionando**
```bash
# Em outro terminal, testar a API
curl http://localhost:8080/health
```

**Resposta esperada:**
```json
{
  "status": "healthy",
  "timestamp": 1705312200,
  "uptime": "0s"
}
```

### 3. **Acessar o Dashboard**
Abra o navegador em: `http://localhost:8080`

Você deve ver:
- ✅ Dashboard com estatísticas
- ✅ Lista de dispositivos (mesmo que vazia)
- ✅ Interface moderna e responsiva

## 🔧 **Teste Completo com Dispositivos**

### 1. **Configurar Dispositivos de Teste**

Primeiro, vamos criar alguns dispositivos de teste no CSV:
```bash
# Editar o arquivo devices.csv
cat > devices.csv << EOF
name,address,username,password
router1,192.168.1.1,admin,password123
router2,192.168.1.2,admin,password456
router3,192.168.1.3,admin,password789
EOF
```

### 2. **Executar Backup com Monitoramento**
```bash
# Executar backup único
./main -config config.json -devices devices.csv

# OU executar em modo daemon
./main -daemon -config config.json -devices devices.csv
```

### 3. **Monitorar em Tempo Real**
```bash
# Em outro terminal, acompanhar métricas
watch -n 5 'curl -s http://localhost:8080/api/v1/stats | jq'
```

## 📊 **Testes da API**

### 1. **Health Check**
```bash
curl http://localhost:8080/health
```

### 2. **Estatísticas Gerais**
```bash
curl http://localhost:8080/api/v1/stats
```

**Resposta esperada:**
```json
{
  "total_devices": 3,
  "healthy_devices": 2,
  "warning_devices": 1,
  "critical_devices": 0,
  "unknown_devices": 0,
  "total_backups": 15,
  "total_failures": 2,
  "average_backup_time": "2m30s"
}
```

### 3. **Métricas de Todos os Dispositivos**
```bash
curl http://localhost:8080/api/v1/metrics
```

### 4. **Métricas de um Dispositivo Específico**
```bash
curl http://localhost:8080/api/v1/metrics/router1
```

### 5. **Alertas Ativos**
```bash
curl http://localhost:8080/api/v1/alerts
```

## 🚨 **Teste de Alertas**

### 1. **Simular Falhas Consecutivas**
```bash
# Criar dispositivos com credenciais inválidas para gerar falhas
cat > devices_test_failures.csv << EOF
name,address,username,password
bad-router1,192.168.99.1,wrong-user,wrong-pass
bad-router2,192.168.99.2,wrong-user,wrong-pass
EOF

# Executar backup (vai falhar)
./main -config config.json -devices devices_test_failures.csv
```

### 2. **Verificar Alertas Gerados**
```bash
# Ver alertas críticos
curl http://localhost:8080/api/v1/alerts | jq '.[] | select(.severity == "critical")'
```

### 3. **Testar Acknowledgment de Alertas**
```bash
# Obter ID do alerta
ALERT_ID=$(curl -s http://localhost:8080/api/v1/alerts | jq -r 'keys[0]')

# Reconhecer alerta
curl -X POST http://localhost:8080/api/v1/alerts/$ALERT_ID

# Verificar se foi reconhecido
curl http://localhost:8080/api/v1/alerts | jq '.[] | select(.acknowledged == true)'
```

## 🎯 **Teste de Cenários Específicos**

### **Cenário 1: Dispositivo Saudável**
```bash
# Configurar dispositivo válido
cat > devices_healthy.csv << EOF
name,address,username,password
healthy-router,192.168.1.100,admin,admin123
EOF

# Executar backup
./main -config config.json -devices devices_healthy.csv

# Verificar status
curl http://localhost:8080/api/v1/metrics/healthy-router
```

### **Cenário 2: Dispositivo com Problemas**
```bash
# Configurar dispositivo com problemas
cat > devices_problematic.csv << EOF
name,address,username,password
slow-router,192.168.1.200,admin,admin123
EOF

# Executar backup múltiplas vezes para gerar alertas
for i in {1..5}; do
  ./main -config config.json -devices devices_problematic.csv
  sleep 10
done
```

### **Cenário 3: Modo Daemon**
```bash
# Iniciar daemon
./main -daemon -config config.json -devices devices.csv

# Em outro terminal, monitorar
watch -n 10 'curl -s http://localhost:8080/api/v1/stats | jq'

# Parar daemon com Ctrl+C
```

## 🔍 **Testes de Integração**

### 1. **Teste com Telegram (se configurado)**
```bash
# Editar config.json para habilitar Telegram
cat > config_telegram.json << EOF
{
    "backup_dir": "backups_mikrotik",
    "schedule": {
        "enabled": true,
        "cron": "0 2 * * *",
        "timezone": "America/Sao_Paulo"
    },
    "telegram": {
        "enabled": true,
        "bot_token": "SEU_BOT_TOKEN",
        "chat_id": "SEU_CHAT_ID",
        "send_zip": true,
        "send_logs": true
    },
    "backup_retention": {
        "enabled": true,
        "keep_days": 30
    }
}
EOF

# Executar com Telegram
./main -config config_telegram.json -devices devices.csv
```

### 2. **Teste de Performance**
```bash
# Medir tempo de resposta da API
time curl http://localhost:8080/api/v1/metrics

# Teste de carga simples
for i in {1..100}; do
  curl -s http://localhost:8080/health > /dev/null
done
```

## 🐛 **Testes de Erro**

### 1. **Porta já em uso**
```bash
# Tentar usar porta já ocupada
./main -api-port 80  # Deve falhar se não for root
```

### 2. **Arquivo de configuração inválido**
```bash
# Criar config inválido
echo "invalid json" > bad_config.json

# Tentar executar
./main -config bad_config.json  # Deve falhar
```

### 3. **CSV malformado**
```bash
# Criar CSV inválido
echo "name,address" > bad_devices.csv  # Sem username/password

# Tentar executar
./main -devices bad_devices.csv  # Deve falhar
```

## 📱 **Teste do Dashboard Web**

### 1. **Interface Responsiva**
- Abrir `http://localhost:8080` no navegador
- Redimensionar janela para testar responsividade
- Verificar se elementos se ajustam

### 2. **Auto-refresh**
- Deixar dashboard aberto
- Executar backups em outro terminal
- Verificar se dados atualizam automaticamente

### 3. **Navegação**
- Testar todos os elementos clicáveis
- Verificar se cores de status estão corretas
- Testar em diferentes navegadores

## 🔧 **Script de Teste Automatizado**

Crie um script para testar tudo automaticamente:

```bash
#!/bin/bash
# test_all.sh

echo "🧪 Iniciando testes do Aegis Backup..."

# Compilar
echo "📦 Compilando..."
go build ./cmd/main.go || exit 1

# Teste básico
echo "🔍 Teste básico..."
./main -config config.json -devices devices.csv &
PID=$!
sleep 3

# Testar API
echo "🌐 Testando API..."
curl -f http://localhost:8080/health || exit 1
curl -f http://localhost:8080/api/v1/stats || exit 1

# Parar servidor
kill $PID

echo "✅ Todos os testes passaram!"
```

## 📊 **Monitoramento Contínuo**

### 1. **Logs em Tempo Real**
```bash
# Acompanhar logs
tail -f /var/log/aegis-backup.log  # Se configurado

# OU usar stdout
./main -daemon 2>&1 | tee aegis.log
```

### 2. **Métricas de Sistema**
```bash
# Monitorar uso de recursos
top -p $(pgrep main)

# Monitorar rede
netstat -tulpn | grep :8080
```

## 🎯 **Checklist de Testes**

- [ ] Compilação sem erros
- [ ] API responde corretamente
- [ ] Dashboard carrega
- [ ] Métricas são coletadas
- [ ] Alertas são gerados
- [ ] Acknowledgment funciona
- [ ] Modo daemon funciona
- [ ] Integração com Telegram (se configurado)
- [ ] Performance aceitável
- [ ] Tratamento de erros

## 🚀 **Próximos Passos**

Após testar tudo:

1. **Configurar dispositivos reais** no `devices.csv`
2. **Configurar Telegram** para notificações
3. **Executar em modo daemon** para produção
4. **Integrar com sistemas de monitoramento** existentes
5. **Configurar alertas personalizados** se necessário

---

**🎉 Agora você tem um sistema completo de backup com monitoramento profissional!**
