# Correções de Segurança e Melhorias - Aegis Backup

Este documento descreve as correções de segurança e melhorias implementadas no projeto Aegis Backup.

## 🔒 Correções de Segurança

### 1. Simplificação SSH - Remoção de Host Key
**Problema**: Complexidade desnecessária com verificação de host key SSH.

**Solução**: 
- Removido campo `host_key` da estrutura `Device`
- Simplificado para usar sempre `ssh.InsecureIgnoreHostKey()`
- Adequado para ambientes controlados e redes confiáveis
- CSV simplificado com apenas 4 colunas obrigatórias

**Arquivos modificados**:
- `internal/config/loader.go` - Estrutura Device simplificada
- `internal/backup/backup.go` - Configuração SSH simplificada

### 2. Permissões de Arquivo
**Problema**: Arquivos de backup com permissões muito permissivas (0644).

**Solução**:
- Alterado permissões de arquivos de backup para 0600 (apenas proprietário)
- Alterado permissões de diretório de backup para 0700 (apenas proprietário)

**Arquivos modificados**:
- `internal/backup/backup.go` - Permissões de arquivo
- `cmd/main.go` - Permissões de diretório

## 🛠️ Melhorias de Robustez

### 3. Validação de Dispositivos
**Problema**: Falta de validação se dispositivos estão configurados corretamente.

**Solução**:
- Validação de lista não vazia de dispositivos
- Validação de campos obrigatórios (name, address, username, password)
- Mensagens de erro específicas para cada problema

**Arquivos modificados**:
- `cmd/main.go` - Validação de dispositivos

### 4. Workers Dinâmicos
**Problema**: Número fixo de workers (5) não era adequado para diferentes quantidades de dispositivos.

**Solução**:
- Implementada função `calculateWorkers()` com lógica inteligente:
  - 1-5 dispositivos: 1 worker por dispositivo
  - 6-10 dispositivos: 5 workers
  - 11+ dispositivos: máximo 8 workers
- Aplicado tanto no modo único quanto no scheduler

**Arquivos modificados**:
- `cmd/main.go` - Função calculateWorkers e uso dinâmico
- `internal/scheduler/scheduler.go` - Mesma lógica no scheduler

### 5. Tratamento de Erros ZIP
**Problema**: Falhas na criação de ZIP quebravam o fluxo de notificações.

**Solução**:
- Melhor tratamento de erros na criação de ZIP
- Continuação da execução mesmo se ZIP falhar
- Logs informativos sobre sucesso/falha
- Limpeza automática de arquivos ZIP vazios

**Arquivos modificados**:
- `internal/archiver/zip.go` - Melhor tratamento de erros
- `cmd/main.go` - Continuação do fluxo em caso de erro
- `internal/scheduler/scheduler.go` - Mesmo tratamento no scheduler

### 6. Timezone nas Mensagens Telegram
**Problema**: Timestamps nas mensagens sempre usavam UTC, ignorando timezone configurado.

**Solução**:
- Funções de formatação agora recebem parâmetro timezone
- Uso do timezone configurado nas mensagens
- Fallback para UTC em caso de timezone inválido
- Exibição do timezone nas mensagens

**Arquivos modificados**:
- `internal/telegram/client.go` - Funções FormatBackupSummary e FormatErrorMessage
- `cmd/main.go` - Passagem do timezone configurado
- `internal/scheduler/scheduler.go` - Mesmo tratamento no scheduler

### 7. Pool de HTTP Clients
**Problema**: Criação de novo HTTP client a cada requisição Telegram.

**Solução**:
- HTTP client reutilizável com configurações otimizadas
- Pool de conexões com MaxIdleConns
- Timeout configurado (30 segundos)
- Reutilização de conexões TCP

**Arquivos modificados**:
- `internal/telegram/client.go` - Client HTTP reutilizável

### 8. Validação de Configuração Robusta
**Problema**: Falta de validação de configuração antes do uso.

**Solução**:
- Validação de diretório de backup
- Validação de expressão cron quando scheduler habilitado
- Validação de timezone
- Validação de token e chat ID do Telegram
- Validação de dias de retenção (1-365)

**Arquivos modificados**:
- `internal/config/loader.go` - Função validateConfig

## 🧪 Testes Unitários

### 9. Testes Básicos Implementados
**Adicionado**:
- Testes para validação de configuração
- Testes para carregamento de dispositivos CSV
- Testes para formatação de mensagens Telegram
- Testes para criação e limpeza de arquivos ZIP

**Arquivos criados**:
- `internal/config/loader_test.go`
- `internal/telegram/client_test.go`
- `internal/archiver/zip_test.go`

## 📋 Arquivos Modificados

### Principais:
- `cmd/main.go` - Validação, workers dinâmicos, permissões
- `internal/config/loader.go` - Host key, validação
- `internal/backup/backup.go` - Segurança SSH, permissões
- `internal/scheduler/scheduler.go` - Workers dinâmicos, timezone
- `internal/telegram/client.go` - HTTP pool, timezone
- `internal/archiver/zip.go` - Tratamento de erros

### Novos:
- `internal/config/loader_test.go`
- `internal/telegram/client_test.go`
- `internal/archiver/zip_test.go`
- `SECURITY_FIXES.md` (este arquivo)

### Atualizados:
- `devices.csv` - Adicionada coluna host_key

## ✅ Status das Correções

- ✅ Vulnerabilidade SSH corrigida
- ✅ Validação de dispositivos implementada
- ✅ Workers dinâmicos implementados
- ✅ Tratamento de erros ZIP melhorado
- ✅ Timezone nas mensagens corrigido
- ✅ Pool de HTTP clients implementado
- ✅ Permissões de arquivo ajustadas
- ✅ Validação de configuração robusta
- ✅ Testes unitários básicos adicionados
- ⏭️ Logs estruturados (opcional, não implementado)

## 🚀 Como Usar

1. **CSV Simplificado**: O arquivo `devices.csv` agora tem apenas 4 colunas obrigatórias:
   ```csv
   name,address,username,password
   router1,192.168.1.1,admin,admin123
   router2,192.168.1.2,admin,admin123
   ```

2. **Validação**: O sistema agora valida automaticamente a configuração antes de iniciar.

3. **Workers**: O número de workers é calculado automaticamente baseado na quantidade de dispositivos.

4. **SSH Simplificado**: Não é necessário configurar host keys SSH - o sistema funciona diretamente com usuário e senha.

O sistema foi simplificado para ser mais fácil de usar em ambientes controlados e redes confiáveis.
