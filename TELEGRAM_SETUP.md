# 🤖 Telegram Setup Guide

Este guia explica como configurar a integração com Telegram para receber notificações e arquivos de backup automaticamente.

## 📋 Pré-requisitos

- Uma conta no Telegram
- Acesso a um grupo/canal onde você quer receber as notificações
- Permissões de administrador no grupo (se aplicável)

## 🤖 Passo 1: Criar um Bot

1. Abra o Telegram e procure por [@BotFather](https://t.me/botfather)
2. Inicie uma conversa e envie `/start`
3. Envie `/newbot` para criar um novo bot
4. Escolha um nome para seu bot (ex: "Aegis Backup Bot")
5. Escolha um username único terminado em "bot" (ex: "aegis_backup_bot")
6. Salve o **token** fornecido (formato: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

## 🆔 Passo 2: Obter o Chat ID

### Para um Grupo:
1. Adicione seu bot ao grupo
2. Envie qualquer mensagem no grupo
3. Acesse: `https://api.telegram.org/bot<SEU_TOKEN>/getUpdates`
4. Procure por `"chat":{"id":-1001234567890}` na resposta
5. O Chat ID será o número negativo (ex: `-1001234567890`)

### Para um Chat Privado:
1. Inicie uma conversa com seu bot
2. Envie qualquer mensagem
3. Acesse a mesma URL acima
4. Procure por `"chat":{"id":123456789}` (número positivo)

## ⚙️ Passo 3: Configurar Permissões

### Para Grupos:
- Certifique-se de que o bot pode enviar mensagens
- Verifique se o bot pode enviar arquivos
- Se necessário, promova o bot a administrador

### Para Canais:
- Adicione o bot como administrador
- Conceda permissão para postar mensagens

## 🔧 Passo 4: Configurar o Aegis Backup

Edite seu `config.json`:

```json
{
    "backup_dir": "backups_mikrotik",
    "schedule": {
        "enabled": true,
        "cron": "0 2 * * *",
        "timezone": "America/Sao_Paulo"
    },
    "telegram": {
        "enabled": true,
        "bot_token": "SEU_TOKEN_AQUI",
        "chat_id": "SEU_CHAT_ID_AQUI",
        "send_zip": true,
        "send_logs": true
    }
}
```

### Opções de Configuração:

- **`enabled`**: `true` para ativar, `false` para desativar
- **`bot_token`**: Token do seu bot (obtido no Passo 1)
- **`chat_id`**: ID do chat/grupo (obtido no Passo 2)
- **`send_zip`**: `true` para enviar arquivos ZIP diários
- **`send_logs`**: `true` para enviar resumos de backup

## ✅ Passo 5: Testar a Configuração

1. Execute o Aegis Backup em modo daemon:
   ```bash
   ./aegis-backup -daemon
   ```

2. Verifique os logs para mensagens como:
   ```
   Telegram client initialized successfully
   ```

3. Se houver erros, verifique:
   - Token do bot está correto
   - Chat ID está correto
   - Bot tem permissões no grupo/canal

## 📱 Tipos de Notificação

### Resumo de Backup:
```
🛡️ Aegis Backup Completed

📊 Summary:
• Devices backed up: 5
• Duration: 2m30s
• ZIP file: backups_2024-01-15.zip
• Date: 2024-01-15 02:00:15

✅ All configurations have been successfully backed up and compressed.
```

### Notificação de Erro:
```
❌ Aegis Backup Error

🔧 Operation: ZIP Creation
⚠️ Error: failed to create ZIP file: permission denied
📅 Time: 2024-01-15 02:00:15

Please check the logs for more details.
```

### Arquivo ZIP:
- Enviado automaticamente após cada backup
- Contém todos os backups do dia
- Nome: `backups_YYYY-MM-DD.zip`

## 🛠️ Solução de Problemas

### Bot não responde:
- Verifique se o token está correto
- Certifique-se de que o bot não foi deletado

### Mensagens não chegam:
- Verifique o Chat ID
- Confirme que o bot está no grupo
- Verifique permissões do bot

### Arquivos não são enviados:
- Limite do Telegram: 50MB por arquivo
- Verifique conexão com internet
- Confirme que `send_zip` está habilitado

### Erro de permissão:
- Bot precisa ser administrador em canais
- Em grupos, verifique se pode enviar mensagens e arquivos

## 🔒 Segurança

- **Nunca compartilhe** seu token do bot
- Use grupos privados para receber backups
- Considere usar um bot dedicado apenas para backups
- Mantenha o `config.json` seguro (já está no .gitignore)

## 📞 Suporte

Se você encontrar problemas:
1. Verifique os logs do Aegis Backup
2. Teste manualmente enviando mensagens para o bot
3. Use o [@BotFather](https://t.me/botfather) para verificar o status do bot
4. Consulte a documentação oficial do Telegram Bot API