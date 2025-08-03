# Rate Limiter em Go

Rate limiter desenvolvido para o desafio técnico da Pós Go - Expert.

## Características

- **Limitação por IP**: Controla requisições baseado no endereço IP do cliente
- **Limitação por Token**: Controla requisições baseado em tokens de acesso (header API_KEY)
- **Precedência de Token**: Configurações de token sobrepõem configurações de IP
- **Storage Plugável**: Suporte a Redis e memória com padrão Strategy
- **Middleware**: Integração fácil como middleware HTTP

## Configuração

### Variáveis de Ambiente

```env
# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Rate Limiter Padrão
IP_RATE_LIMIT=10
IP_BLOCK_TIME=300
TOKEN_RATE_LIMIT=100
TOKEN_BLOCK_TIME=300

# Servidor
SERVER_PORT=8080

# Tokens Específicos (formato: TOKEN_nome=limite:tempo_bloqueio)
TOKEN_abc123=100:60
TOKEN_xyz789=50:120
TOKEN_premium=1000:30
```

### Endpoints

- GET / - Endpoint principal
- GET /health - Health check
- GET /test - Endpoint de teste


## Como Testar

#### com Curl

- Teste básico
curl http://localhost:8080/

- Teste com token
curl -H "API_KEY: abc123" http://localhost:8080/

- Teste rate limiting (múltiplas vezes)
for i in {1..15}; do curl http://localhost:8080/ && echo; done

#### com Postman

Cadastrar os tokens nos headers e fazer as requisições com e sem os tokens ativos:
