# TransactionHub (Desafio Técnico)

Sistema de processamento de transações financeiras construído em **Go**, implementando um núcleo transacional robusto para operações de crédito, débito, reserva, captura, estorno e transferência entre contas.

Este projeto foi construído focando nos requisitos de consistência, concorrência, idempotência e mensageria assíncrona exigidos em desafios técnicos de gateways de pagamento e fintechs (como a PagueVeloz).

---

## Decisões Técnicas e Arquiteturais

A implementação foi realizada em **Go (Golang)** como uma decisão técnica estratégica para alta performance:

- **Performance e concorrência nativa**: Go possui goroutines e primitivas de sincronização (`sync.Mutex`, `sync.Map`) embutidas na linguagem, tornando o controle de concorrência muito mais simples e eficiente para um sistema de alta volumetria transacional.
- **Binário único e leve**: O binário compilado em Go é extremamente enxuto, facilitando o deploy num container Docker e sem necessidade do framework runtime.
- **Adequação ao contexto Fintech**: O mercado financeiro de alta precisão (ex: gateways de pagamento) frequentemente adota Go pela baixíssima e previsível latência no processamento.

### Arquitetura (Clean Architecture, DDD e CQRS)

O projeto segue os princípios de **Clean Architecture**, **DDD (Domain-Driven Design)** e uma base de **CQRS (Command Query Responsibility Segregation)**, entregando uma clara separação de conceitos e abstração por Injeção de Dependências. Essa modelagem facilita a testabilidade da regra de negócios e o desacoplamento necessário para evoluir o sistema para microsserviços.

A estrutura de diretórios foi diretamente moldada por esses padrões:

```text
internal/
├── domain/          # (DDD) Coração do software: Entidades Ricas (Account, Transaction) com regras de negócio blindadas.
├── application/     # (CQRS) Orquestração de casos de uso (Command Handlers).
│   ├── commands/    # Ações que mutam o estado (Credit, Debit, Transfer, Reversal).
│   └── dto/         # Objetos de transferência para retorno isolado.
├── infrastructure/  # Adaptadores e Detalhes Externos (PostgreSQL via GORM, Simulador Event Publisher).
├── api/             # Camada de Apresentação (Controllers HTTP configurados via Gin Gonic).
└── boostrap/        # Ponto central de Composição (IoC / Injeção de Dependências manual).
```

- **DDD (Domain-Driven Design):** Encardido no pacote `domain`. Usamos entidades com comportamento encapsulado (*Rich Domain Model*). Operações como debito e limite são executadas diretamente nos métodos das entidades (ex: `acc.Debit()`), eliminando o antipadrão de *Modelos Anêmicos*. Todo o sistema orbita ao redor e depende desse núcleo (Inversão de Dependência).
- **CQRS (Segregação de Responsabilidade):** Aplicado na camada `application/commands`. Fluxos que modificam o estado transacional possuem seus próprios manipuladores de casos de uso dedicados (`CreditHandler`, `ReversalHandler`, etc.). Em vez de usar grandes "Services" inchados, isolamos a intenção de mutação de estado (Commands) o que favorece manutenibilidade paralela.
- **Clean Architecture:** Como efeito consequente, a comunicação HTTP (Gin) e a manipulação relacional de dados (GORM) são tratadas meramente como plugins ou *Detalhes da Infraestrutura*.

### Controle de Concorrência e Race Conditions

Para evitar que duas transações simultâneas gerem anomalias no saldo de uma mesma conta, o sistema implementa **locks em memória (Pessimistic Locking)** usando `sync.Map` e `sync.Mutex`:

- `LockAccount(accountID)` — Adquire o lock unicamente para operações em uma conta (ex: Crédito).
- `LockTransferAccounts(sourceID, destinationID)` — Trava **ambas as contas** na transação de Transferência, porém garantindo que a aquisição do lock ocorra em **ordem alfabética**. Isso previne completamente a ocorrência do cenário clássico de **Deadlock**.

### Publicação Assíncrona de Eventos (Eventual Consistency & Retry)

Para não onerar o fluxo transacional síncrono com a necessidade de notificar outros sistemas (mensageria externa), foi implementada uma interface `EventPublisher`.
A implementação `LogEventPublisher` **simula a publicação em uma fila externa** (podendo futuramente ser RabbitMQ/Kafka):

- O worker roda em paralelo (goroutine).
- Implementa um fallback de **Exponential Backoff**: Em caso de falha sistêmica simulada, o sistema aguarda um intervalo que dobra a cada tentativa, garantindo resiliência sem congestionar serviços de terceiros momentaneamente instáveis.

### Idempotência

Operações duplicadas criadas via duplo-click pelo frontend ou retentativas de rede, não afetam o saldo da conta. Todas as operações recebem e validam um `reference_id` (com chave `UNIQUE` no banco de dados). Caso esse `reference_id` transacional já tenha sido processado e retornado sucesso, a API o devolve imediatamente, **sem repetição e sem alterar o banco**.

---

## Estrutura e Ferramentas Utilizadas

| Ferramenta | Descrição e Propósito |
|---|---|
| **Go 1.22+** | Linguagem principal em que os pacotes foram escritos (sem cgo dependency) |
| **Gin Gonic** | Web framework HTTP rápido e focado com roteamento e middlewares. |
| **GORM** | ORM robusto em Go (Database-first e Code-first) usado para mapeamento Postgres. |
| **PostgreSQL 16** | Um SGBD relacional moderno com bloqueios ACID e Constraints transacionais perfeitos. |
| **Swagger / Swaggo** | Para documentação unificada e exploração automatizada da API localmente. |
| **Docker & Compose** | Containerização agnóstica provendo um ambiente 100% plug & play transparente. |

---

## Pré-requisitos e Execução

### Opção 1: Via Docker Compose (Recomendado)

Sobe toda a infraestrutura (API e PostgreSQL) automaticamente em instantes:

```bash
docker-compose up -d --build
```
> A aplicação web responderá na porta `8080`, e o banco de dados rodará na `5432` inicializando as tabelas automaticamente via bootscript presente em `docker/init.sql`.

### Opção 2: Localmente via Golang Nativo

1. Defina as variáveis seguindo o modelo do `.envexample`, no console local ou crie um arquivo `.env`:
```bash
export DB_HOST=YOURHOST
export DB_USER=YOUR_DBUSER
export DB_PASSWORD=YOUR_DBPASSWORD
export DB_NAME=YOUR_DBNAME
export DB_PORT=YOUR_PORT
```
2. Baixe os recursos e rode o binário do código fonte:
```bash
go run ./cmd/api/main.go
```

---

## Documentação Interativa da API (Swagger)

A API conta com interface gráfica Swagger UI integrada automaticamente na engine web do framework.

Com o sistema rodando, acesse em seu navegador:
👉 **[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

Lá você poderá verificar, testar e enviar de modo simples os payloads reais baseados nos DTOs exatos da api sob a porta :8080.

---

## Principais Endpoints e Uso

O sistema gerencia todas as suas operações no arquivo raiz através da entidade Account e das Transactions ligadas a ela.

### 1. Criar uma Conta (`POST /accounts`)

```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "id": "ACC-001",
    "client_id": "CLIENT-001",
    "balance": 0,
    "credit_limit": 50000,
    "status": "active"
  }'
```

### 2. Fluxos Transacionais (`POST /transactions`)

Todas as requisições passam pela Controller de transações. As variáveis se moldam conforme o valor no parametro **`operation`**.

#### Tipos Base Suportados:

**Crédito (`credit`):** Adição de saldo simples e puro.
```json
{
  "operation": "credit",
  "account_id": "ACC-001",
  "amount": 2000,
  "currency": "BRL",
  "reference_id": "DP-001"
}
```

**Débito (`debit`):** Remove o saldo da conta, verificando primeiramente por (available_balance + credit_limit).
```json
{
  "operation": "debit",
  "account_id": "ACC-001",
  "amount": 500,
  "currency": "BRL",
  "reference_id": "WD-001"
}
```

**Transferência (`transfer`):** Retira do saldo da conta origem validando limites da mesma regra de débito, travando e creditando a quantia exata a conta de destino simultaneamente. (Requer uso do ID extra)
```json
{
  "operation": "transfer",
  "account_id": "ACC-001",
  "destination_account_id": "ACC-002",
  "amount": 1000,
  "currency": "BRL",
  "reference_id": "TRFR-001"
}
```

**Estorno (`reversal`):** Desfaz logicamente a transação original enviada. Credita casos de crédito devolvendo debitos.
```json
{
  "operation": "reversal",
  "account_id": "ACC-001",
  "original_transaction_id": "seu-transaction-id-origial",
  "amount": 500,
  "currency": "BRL",
  "reference_id": "REV-001"
}
```

**Reserva (`reserve`):** (Opcional - prende temporariamente um saldo na conta, retirando-o do fluxo principal disponível)
```json
{
  "operation": "reserve",
  "account_id": "ACC-001",
  "amount": 200,
  "currency": "BRL",
  "reference_id": "RES-001"
}
```

**Captura (`capture`):** (Opcional - Confirma fisicamente que o recurso reservado pode ser utilizado/transacionado)
```json
{
  "operation": "capture",
  "account_id": "ACC-001",
  "amount": 200,
  "currency": "BRL",
  "reference_id": "CAP-001"
}
```

---

## Bateria de Testes (Unitário / Fluxos de Raiz)

A aplicação garante sua resiliência rodando centenas de combinações em cobertura, isolando a regra da infra.

```bash
# Rodar todos os relatórios unitários (Handlings e regras)
go test ./... -v

# Realizar checagens rígidas de colisão/vazamento/simultaniedade no Mutex (Race Condition Detect):
go test -race ./...
```
