# Omni Infrastructure Provider for Proxmox

Can be used to automatically provision Talos nodes in a Proxmox cluster.

## Requirements

- Proxmox VE cluster
- User account with sufficient permissions to manage VMs and resources (example uses root)
- Omni account and infrastructure provider key
- Network connectivity between the infrastructure provider and your Proxmox cluster

## Running Infrastructure Provider

Create the configuration file for the provider:

```yaml
proxmox:
  username: root
  password: 123456
  url: "https://homelab.proxmox:8006/api2/json"
  insecureSkipVerify: true
  realm: "pam"
```

> **Note:**
>
> - Replace the `url` value with the address of your own Proxmox server.
> - You can use a different user instead of `root` if you grant it the necessary permissions to manage resources in your Proxmox cluster.

### Using Docker

> **Note:** The `--omni-service-account-key` flag expects an *infra provider key*, not an Omni service account key.
> Make sure to provide the correct key type.

Run the provider using Docker:

```bash
docker run -it -d \
  -v ./config.yaml:/config.yaml \
  ghcr.io/siderolabs/omni-infra-provider-proxmox \
  --config-file /config.yaml \
  --omni-api-endpoint https://<account-name>.omni.siderolabs.io/ \
  --omni-service-account-key <infra-provider-key>
```

### Example Docker Compose

You can also run the provider using Docker Compose.
Create a `docker-compose.yaml` file:

```yaml
services:
  omni-infra-provider-proxmox:
    image: ghcr.io/siderolabs/omni-infra-provider-proxmox
    volumes:
      - ./config.yaml:/config.yaml
    command: >
      --config-file /config.yaml
      --omni-api-endpoint https://<account-name>.omni.siderolabs.io/
      --omni-service-account-key <infrastructure-provider-key>
    restart: unless-stopped
```

Start the provider:

```bash
docker compose up -d
```

## Creating a Machine Class for Auto Provision

To enable automatic provisioning of Talos nodes, you need to define a machine class of type `auto-provision` in Omni.
This class specifies the configuration for new VMs, such as CPU, memory, and disk size.

Example machine class definition:

```yaml
apiVersion: infrastructure.omni.siderolabs.io/v1alpha1
kind: MachineClass
metadata:
  name: proxmox-auto
spec:
  type: auto-provision
  provider: proxmox
  config:
    cpu: 4
    memory: 8192 # in MB
    diskSize: 40 # in GB
    # Add other Proxmox-specific options as needed
```

Apply the machine class to your Omni account using the Omni UI or CLI.

### Scaling a Cluster with the Machine Class

You can now use above `proxmox-auto` machine class to scale an existing cluster up or down, or to create a new cluster:

- **To scale up:** Increase the desired number of machines in your cluster configuration.
  Omni will automatically provision new VMs using the specified machine class.
- **To scale down:** Decrease the desired number of machines.
  Omni will remove excess VMs accordingly.
- **To create a new cluster:** Specify the machine class in your cluster manifest when creating a new cluster.

Example cluster manifest snippet:

```yaml
spec:
  machineClass: proxmox-auto
  replicas: 3
```

### Storage Selector Requirement During VM Sync

> **Note:**
> During the `vmSync` step, you may encounter an error requiring a Storage Selector.
> This is a CEL (Common Expression Language) expression used to select the appropriate Proxmox storage for VM disk images.
>
> To resolve this, add a `storageSelector` field to your machine class configuration.

```yaml
config:
  ...
  storageSelector: 'name == "local-lvm"'
```

Replace `"local-lvm"` with the name of the storage you want to use for VM disks in your Proxmox cluster.

### Local Development with Docker

For a quick local build and run without needing buildx or the full kres toolchain:

1. **Create configuration files:**

   ```bash
   cp config.yaml.example config.yaml
   cp .env.example .env
   ```

2. **Edit `config.yaml`** with your Proxmox server URL and credentials.

3. **Edit `.env`** with your Omni API endpoint and infrastructure provider key.

4. **Build and start the provider:**

   ```bash
   docker compose up --build
   ```

   To run in the background:

   ```bash
   docker compose up --build -d
   ```

5. **View logs:**

   ```bash
   docker compose logs -f
   ```

6. **Stop the provider:**

   ```bash
   docker compose down
   ```

### Local Docker Mode (Development without Proxmox)

The `--local` flag replaces the Proxmox backend with a local Docker-based provisioner. Instead of creating VMs in Proxmox, it creates containers on your local Docker daemon. This is useful for development and testing when you don't have access to a Proxmox cluster.

**Requirements:**
- Docker running locally (Docker Desktop, colima, etc.)
- Omni API endpoint and infrastructure provider key (still required to register as a provider with Omni)

**Using Docker Compose:**

1. Set up your `.env` file with Omni credentials:

   ```bash
   cp .env.example .env
   # Edit .env with your OMNI_API_ENDPOINT and OMNI_SERVICE_ACCOUNT_KEY
   ```

2. Build and start in local mode:

   ```bash
   docker compose --profile local up --build local
   ```

   Or in the background:

   ```bash
   docker compose --profile local up --build local -d
   ```

**Using Docker directly:**

```bash
docker build -f Dockerfile.local -t omni-provider-local .

docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  omni-provider-local \
  --local \
  --omni-api-endpoint https://<account-name>.omni.siderolabs.io/ \
  --omni-service-account-key <infra-provider-key>
```

**Using the binary directly:**

```bash
go build -o provider ./cmd/omni-infra-provider-proxmox

./provider --local \
  --omni-api-endpoint https://<account-name>.omni.siderolabs.io/ \
  --omni-service-account-key <infra-provider-key>
```

> **Note:** No `config.yaml` is needed in local mode — the Proxmox configuration is not used.

### Dry-Run Mode

The `--dry-run` flag lets you validate your Proxmox configuration and test connectivity without connecting to Omni. It reads `config.yaml`, connects to the Proxmox API, and reports the version and available nodes.

```bash
# Using Docker Compose
docker compose --profile test up --build dry-run

# Using the binary
./provider --dry-run --config-file config.yaml
```

### CLI Flags Reference

| Flag | Default | Description |
|------|---------|-------------|
| `--config-file` | | Proxmox provider config file path |
| `--omni-api-endpoint` | `$OMNI_ENDPOINT` | Omni API endpoint URL |
| `--omni-service-account-key` | `$OMNI_SERVICE_ACCOUNT_KEY` | Omni infrastructure provider key |
| `--provider-name` | `Proxmox` | Provider name as it appears in Omni |
| `--provider-description` | `Proxmox infrastructure provider` | Provider description in Omni |
| `--id` | `proxmox` | Infra provider ID for resource matching |
| `--insecure-skip-verify` | `false` | Skip TLS verification for Omni |
| `--dry-run` | `false` | Validate config and test Proxmox connectivity |
| `--local` | `false` | Use local Docker backend instead of Proxmox |

### Using Executable

Build the project (should have docker and buildx installed):

```bash
make omni-infra-provider-linux-amd64
```

Run the executable:

```bash
_out/omni-infra-provider-linux-amd64 --config config.yaml --omni-api-endpoint https://<account-name>.o2b.com.br/ --omni-service-account-key <service-account-key>
```

---

# Omni Infrastructure Provider for Proxmox (PT-BR Version)

Pode ser utilizado para provisionar automaticamente nós Talos em um cluster Proxmox.

## Requisitos

- Cluster Proxmox VE
- Conta de usuário com permissões suficientes para gerenciar VMs e recursos (o exemplo usa root)
- Conta Omni e chave de provedor de infraestrutura
- Conectividade de rede entre o provedor de infraestrutura e o seu cluster Proxmox

## Executando o Provedor de Infraestrutura

Crie o arquivo de configuração para o provedor:

```yaml
proxmox:
  username: root
  password: 123456
  url: "https://homelab.proxmox:8006/api2/json"
  insecureSkipVerify: true
  realm: "pam"
```

> **Nota:**
>
> - Substitua o valor de `url` pelo endereço do seu próprio servidor Proxmox.
> - Você pode utilizar um usuário diferente de `root` caso conceda a ele as permissões necessárias para gerenciar recursos no seu cluster Proxmox.

### Usando Docker

> **Nota:** A flag `--omni-service-account-key` espera uma *chave de provedor de infraestrutura*, não uma chave de conta de serviço Omni.
> Certifique-se de fornecer o tipo de chave correto.

Execute o provedor usando Docker:

```bash
docker run -it -d \
  -v ./config.yaml:/config.yaml \
  ghcr.io/siderolabs/omni-infra-provider-proxmox \
  --config-file /config.yaml \
  --omni-api-endpoint https://<nome-da-conta>.omni.siderolabs.io/ \
  --omni-service-account-key <chave-do-provedor-de-infra>
```

### Exemplo com Docker Compose

Você também pode executar o provedor usando Docker Compose.
Crie um arquivo `docker-compose.yaml`:

```yaml
services:
  omni-infra-provider-proxmox:
    image: ghcr.io/siderolabs/omni-infra-provider-proxmox
    volumes:
      - ./config.yaml:/config.yaml
    command: >
      --config-file /config.yaml
      --omni-api-endpoint https://<nome-da-conta>.omni.siderolabs.io/
      --omni-service-account-key <chave-do-provedor-de-infraestrutura>
    restart: unless-stopped
```

Inicie o provedor:

```bash
docker compose up -d
```

## Criando uma Classe de Máquina para Provisionamento Automático

Para habilitar o provisionamento automático de nós Talos, você precisa definir uma classe de máquina do tipo `auto-provision` no Omni.
Essa classe especifica a configuração para novas VMs, como CPU, memória e tamanho do disco.

Exemplo de definição de classe de máquina:

```yaml
apiVersion: infrastructure.omni.siderolabs.io/v1alpha1
kind: MachineClass
metadata:
  name: proxmox-auto
spec:
  type: auto-provision
  provider: proxmox
  config:
    cpu: 4
    memory: 8192 # em MB
    diskSize: 40 # em GB
    # Adicione outras opções específicas do Proxmox conforme necessário
```

Aplique a classe de máquina à sua conta Omni usando a UI ou a CLI do Omni.

### Escalando um Cluster com a Classe de Máquina

Agora você pode usar a classe de máquina `proxmox-auto` acima para escalar um cluster existente para cima ou para baixo, ou para criar um novo cluster:

- **Para escalar para cima:** Aumente o número desejado de máquinas na configuração do seu cluster.
  O Omni provisionará automaticamente novas VMs usando a classe de máquina especificada.
- **Para escalar para baixo:** Diminua o número desejado de máquinas.
  O Omni removerá as VMs excedentes automaticamente.
- **Para criar um novo cluster:** Especifique a classe de máquina no manifesto do cluster ao criá-lo.

Exemplo de trecho de manifesto de cluster:

```yaml
spec:
  machineClass: proxmox-auto
  replicas: 3
```

### Requisito de Seletor de Armazenamento Durante a Sincronização de VMs

> **Nota:**
> Durante a etapa `vmSync`, você pode encontrar um erro exigindo um Seletor de Armazenamento.
> Trata-se de uma expressão CEL (Common Expression Language) usada para selecionar o armazenamento Proxmox adequado para imagens de disco de VMs.
>
> Para resolver isso, adicione o campo `storageSelector` à configuração da sua classe de máquina.

```yaml
config:
  ...
  storageSelector: 'name == "local-lvm"'
```

Substitua `"local-lvm"` pelo nome do armazenamento que deseja usar para os discos de VM no seu cluster Proxmox.

### Desenvolvimento Local com Docker

Para uma build e execução local rápidas, sem precisar do buildx ou do toolchain completo do kres:

1. **Crie os arquivos de configuração:**

   ```bash
   cp config.yaml.example config.yaml
   cp .env.example .env
   ```

2. **Edite o `config.yaml`** com a URL e as credenciais do seu servidor Proxmox.

3. **Edite o `.env`** com o endpoint da API Omni e a chave do provedor de infraestrutura.

4. **Build e inicialização do provedor:**

   ```bash
   docker compose up --build
   ```

   Para executar em segundo plano:

   ```bash
   docker compose up --build -d
   ```

5. **Visualizar logs:**

   ```bash
   docker compose logs -f
   ```

6. **Parar o provedor:**

   ```bash
   docker compose down
   ```

### Modo Local Docker (Desenvolvimento sem Proxmox)

A flag `--local` substitui o backend Proxmox por um provisionador baseado em Docker local. Em vez de criar VMs no Proxmox, ele cria contêineres no seu Docker local. Útil para desenvolvimento e testes quando você não tem acesso a um cluster Proxmox.

**Requisitos:**
- Docker rodando localmente (Docker Desktop, colima, etc.)
- Endpoint da API Omni e chave do provedor de infraestrutura (ainda necessários para registro como provedor no Omni)

**Usando Docker Compose:**

1. Configure o arquivo `.env` com as credenciais do Omni:

   ```bash
   cp .env.example .env
   # Edite .env com OMNI_API_ENDPOINT e OMNI_SERVICE_ACCOUNT_KEY
   ```

2. Build e execução em modo local:

   ```bash
   docker compose --profile local up --build local
   ```

   Ou em segundo plano:

   ```bash
   docker compose --profile local up --build local -d
   ```

**Usando Docker diretamente:**

```bash
docker build -f Dockerfile.local -t omni-provider-local .

docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  omni-provider-local \
  --local \
  --omni-api-endpoint https://<nome-da-conta>.omni.siderolabs.io/ \
  --omni-service-account-key <chave-do-provedor-de-infra>
```

**Usando o binário diretamente:**

```bash
go build -o provider ./cmd/omni-infra-provider-proxmox

./provider --local \
  --omni-api-endpoint https://<nome-da-conta>.omni.siderolabs.io/ \
  --omni-service-account-key <chave-do-provedor-de-infra>
```

> **Nota:** Nenhum `config.yaml` é necessário no modo local — a configuração do Proxmox não é utilizada.

### Modo Dry-Run

A flag `--dry-run` permite validar a configuração do Proxmox e testar a conectividade sem se conectar ao Omni. Ela lê o `config.yaml`, conecta à API do Proxmox e reporta a versão e os nós disponíveis.

```bash
# Usando Docker Compose
docker compose --profile test up --build dry-run

# Usando o binário
./provider --dry-run --config-file config.yaml
```

### Referência de Flags da CLI

| Flag | Padrão | Descrição |
|------|--------|-----------|
| `--config-file` | | Caminho do arquivo de configuração Proxmox |
| `--omni-api-endpoint` | `$OMNI_ENDPOINT` | URL do endpoint da API Omni |
| `--omni-service-account-key` | `$OMNI_SERVICE_ACCOUNT_KEY` | Chave do provedor de infraestrutura Omni |
| `--provider-name` | `Proxmox` | Nome do provedor exibido no Omni |
| `--provider-description` | `Proxmox infrastructure provider` | Descrição do provedor no Omni |
| `--id` | `proxmox` | ID do provedor de infra para correspondência de recursos |
| `--insecure-skip-verify` | `false` | Ignora verificação TLS para o Omni |
| `--dry-run` | `false` | Valida configuração e testa conectividade Proxmox |
| `--local` | `false` | Usa backend Docker local em vez do Proxmox |

### Usando o Executável

Faça o build do projeto (é necessário ter docker e buildx instalados):

```bash
make omni-infra-provider-linux-amd64
```

Execute o executável:

```bash
_out/omni-infra-provider-linux-amd64 --config config.yaml --omni-api-endpoint https://<nome-da-conta>.o2b.com.br/ --omni-service-account-key <chave-da-conta-de-servico>
```
GenAI4Cloud inc. and Sm4rt.Works Ltda. @US @UAE @EU @LATAM